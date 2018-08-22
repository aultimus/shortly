package shortly

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/aultimus/shortly/db"
	"github.com/cocoonlife/timber"
	"github.com/gorilla/mux"
)

const (
	LenShortened  = 6
	JSONMimeType  = "application/json"
	ContentType   = "Content-Type"
	maxCollisions = 64
)

type App struct {
	server *http.Server
	store  db.DBer
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init(store db.DBer) error {
	router := mux.NewRouter()

	router.HandleFunc("/", a.RootHandler).Methods(http.MethodGet)

	router.HandleFunc("/health",
		a.HealthHandler).Methods(http.MethodGet)

	router.HandleFunc("/create",
		a.CreateHandler).Methods(http.MethodGet)

	router.HandleFunc("/v1/create",
		a.CreateJSONHandler).Methods(http.MethodPost)

	// The restful way to do it would be to GET /url/id but that is a bit of a longer string,
	// we could redirect to that url?
	router.HandleFunc("/v1/redirect/{url}",
		a.RedirectJSONHandler).Methods(http.MethodGet)

	router.HandleFunc("/{url}",
		a.RedirectHandler).Methods(http.MethodGet)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	a.server = server
	a.store = store

	return nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentType, "text/html")
	// TODO: don't do everytime
	b, err := ioutil.ReadFile("static/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		timber.Errorf(err.Error())
		return
	}
	_, err = w.Write(b)
	if err != nil {
		timber.Errorf(err.Error())
	}
}

func (a *App) HealthHandler(w http.ResponseWriter, r *http.Request) {
	// pass - implicit 200 OK
}

type RedirectResponse struct {
	OriginalURL string `json:"original_url"`
	Err         string `json:"error"`
}

// TODO: refactor out common code amongst Redirect* handlers
// TODO: Test this handler
func (a *App) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set(ContentType, "text/html")

	shortenedURL := vars["url"]
	if shortenedURL == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	timber.Infof("Handling request for shortened url %s", shortenedURL)

	storedURL, err := a.store.Get(shortenedURL)
	if err != nil {
		switch err.(type) {
		case *db.ErrDB:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// else assume a Not found error (could declare this error type and switch on it)
		w.WriteHeader(http.StatusNotFound)
		timber.Errorf(err.Error())
	} else {
		http.Redirect(w, r, storedURL.OriginalURL, http.StatusMovedPermanently)
	}
}

// RedirectJSONHandler handles GET access to shortened urls, this endpoint is publically available
// curl http://localhost:8080/v1/redirect/foo
// TODO: respond with correct status codes
func (a *App) RedirectJSONHandler(w http.ResponseWriter, r *http.Request) {
	resp := &RedirectResponse{}
	vars := mux.Vars(r)
	w.Header().Set(ContentType, JSONMimeType)

	shortenedURL := vars["url"]
	if shortenedURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
	timber.Infof("Handling JSON request for shortened url %s", shortenedURL)

	storedURL, err := a.store.Get(shortenedURL)
	if err != nil {
		switch err.(type) {
		case *db.ErrDB:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// else assume a Not found error (could declare this error type and switch on it)
		w.WriteHeader(http.StatusNotFound)
		resp.Err = err.Error()
	} else {
		resp.OriginalURL = storedURL.OriginalURL
	}

	b, err := json.Marshal(resp)
	if err != nil {
		timber.Errorf("failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

// CreateHandler provides the create functionality for the website
func (a *App) CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentType, "text/html")
	err := r.ParseForm()
	if err != nil {
		timber.Errorf(err.Error())
	}
	originalURL := r.Form.Get("url")

	shortenedURL, err := a.Create(&CreateRequest{originalURL}, &MD5Hash{})
	if err != nil {
		switch err.(type) {
		case *db.ErrCollision:
			timber.Errorf(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
			timber.Errorf(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// TODO: write proper html
	w.Write([]byte(shortenedURL))
}

type CreateRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateResponse struct {
	ShortenedURL string `json:"shortened_url"`
	Err          string `json:"error"`
}

// CreateJSONHandler handles the creation of new shortened URLS
// curl localhost:8080/v1/create -d '{"original_url": "http://foobarcat.blogspot.com"}'
// TODO: Add test case where mandatory field original_url is missing
func (a *App) CreateJSONHandler(w http.ResponseWriter, r *http.Request) {
	resp := &CreateResponse{}
	w.Header().Set(ContentType, JSONMimeType)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		timber.Errorf("create handler failed to read bytes: %s", err.Error())
		resp.Err = err.Error()
		b, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	req := &CreateRequest{}
	err = json.Unmarshal(b, req)
	if err != nil {
		timber.Errorf("create handler failed to unmarshal bytes: %s", string(b))
		resp.Err = err.Error()
		b, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	shortenedURL, err := a.Create(req, &MD5Hash{})
	if err != nil {
		switch err.(type) {
		case *db.ErrCollision:
			timber.Errorf(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
			timber.Errorf(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	resp = &CreateResponse{ShortenedURL: shortenedURL}
	if err != nil {
		resp.Err = err.Error()
	}
	b, err = json.Marshal(resp)
	if err != nil {
		timber.Errorf("failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (a *App) doCreate(originalURL string, permutedValue string, hasher Hasher) (string, error) {
	shortenedURL := hasher.Hash(permutedValue)
	timber.Infof("Create request for [%s], hashes to [%s]", permutedValue, shortenedURL)
	// lets try and store it

	storedURL, err := a.store.Get(shortenedURL)
	if err == nil {
		// check if data is equal
		if storedURL.OriginalURL == originalURL {
			return shortenedURL, nil
		}

		// collision
		return "", db.NewErrCollision(fmt.Sprintf("key [%s] already exists", shortenedURL))
	}
	switch err.(type) {
	case *db.ErrNotFound:
		// pass
	default:
		return "", err
	}

	storedURL = &db.StoredURL{permutedValue}
	err = a.store.Create(shortenedURL, storedURL)
	return shortenedURL, err
}

func (a *App) Create(req *CreateRequest, hasher Hasher) (string, error) {
	// attempt to generate hash and store without permutation
	shortenedURL, err := a.doCreate(req.OriginalURL, req.OriginalURL, hasher)
	if err == nil {
		// success
		return shortenedURL, err
	}

	switch err.(type) {
	case *db.ErrCollision:
		// probably a collision - todo refine this assumption
		// pass
	default:
		return shortenedURL, err
	}

	// permute in case of collision
	for i := 0; i < maxCollisions; i++ {
		suffix := strconv.Itoa(i)
		newValue := req.OriginalURL + suffix
		shortenedURL, err := a.doCreate(req.OriginalURL, newValue, hasher)
		if err == nil {
			// success
			return shortenedURL, err
		}

		switch err.(type) {
		case *db.ErrCollision:
			// probably a collision - todo refine this assumption
			continue
		default:
			return shortenedURL, err
		}
	}
	return "", db.NewErrCollision(fmt.Sprintf("failed to store %s, too many collisions", req.OriginalURL))
}

type Hasher interface {
	Hash(string) string
}

type MD5Hash struct {
}

func (h *MD5Hash) Hash(in string) string {
	hash := md5.Sum([]byte(in))
	s := base64.URLEncoding.EncodeToString((hash[:]))
	return s[0:LenShortened]
}
