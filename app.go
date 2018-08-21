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

	router.HandleFunc("/create",
		a.CreateJSONHandler).Methods(http.MethodPost)

	// The restful way to do it would be to GET /url/id but that is a bit of a longer string,
	// we could redirect to that url?
	router.HandleFunc("/redirect/{url}",
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
	fmt.Printf("Handling request for shortened url %s\n", shortenedURL)

	storedURL, err := a.store.Get(shortenedURL)
	if err != nil {
		switch err.(type) {
		case *db.ErrDB:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// else assume a Not found error (could declare this error type and switch on it)
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err.Error())
	} else {
		http.Redirect(w, r, storedURL.OriginalURL, http.StatusMovedPermanently)
	}
}

// RedirectJSONHandler handles GET access to shortened urls, this endpoint is publically available
// curl http://localhost:8080/redirect/foo
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
	fmt.Printf("Handling JSON request for shortened url %s\n", shortenedURL)

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
		fmt.Printf("failed to marshal response\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

type CreateRequest struct {
	OriginalURL   string `json:"original_url"`
	NumCollisions int    `json:"num_collisions"` // debug option
}

type CreateResponse struct {
	ShortenedURL string `json:"shortened_url"`
	Err          string `json:"error"`
}

// CreateJSONHandler handles the creation of new shortened URLS
// curl localhost:8080/create -d '{"original_url": "http://foobarcat.blogspot.com"}'
// TODO: Add test case where mandatory field original_url is missing
func (a *App) CreateJSONHandler(w http.ResponseWriter, r *http.Request) {
	resp := &CreateResponse{}
	w.Header().Set(ContentType, JSONMimeType)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("create handler failed to read bytes: %s\n", err.Error())
		resp.Err = err.Error()
		b, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	req := &CreateRequest{}
	err = json.Unmarshal(b, req)
	if err != nil {
		fmt.Printf("create handler failed to unmarshal bytes: %s\n", string(b))
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
			// try again rather than error - TODO
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
			fmt.Println(err.Error())
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
		fmt.Printf("failed to marshal response\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (a *App) doCreate(originalURL string, permutedValue string, hasher Hasher) (string, error) {
	shortenedURL := hasher.Hash(permutedValue)
	fmt.Printf("Create request for [%s], hashes to [%s]\n", permutedValue, shortenedURL)
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
	// TODO: reject empty string
	// attempt to generate hash and store without permutation
	fmt.Printf("num simulated collisions %d\n", req.NumCollisions)
	if req.NumCollisions > 0 {
		hasher = &CollisionHash{maxCollisions: req.NumCollisions}
	}
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

type CollisionHash struct {
	maxCollisions int
	numCollisions int
}

func (c *CollisionHash) Hash(in string) string {
	if c.numCollisions > c.maxCollisions {
		m := &MD5Hash{}
		return m.Hash(in)
	}
	c.numCollisions++
	return "foo"
}
