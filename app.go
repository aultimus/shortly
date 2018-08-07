package shortly

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aultimus/shortly/db"
	"github.com/gorilla/mux"
)

const (
	LenShortened = 6
	JSONMimeType = "application/json"
	ContentType  = "Content-Type"
)

type App struct {
	server *http.Server
	store  db.DBer
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init() error {
	router := mux.NewRouter()

	router.HandleFunc("/create",
		a.CreateHandler).Methods(http.MethodPost)

	// The restful way to do it would be to GET /url/id but that is a bit of a longer string,
	// we could redirect to that url?
	router.HandleFunc("/{url}",
		a.RedirectHandler).Methods(http.MethodGet)

	router.HandleFunc("/{url}",
		a.DeleteHandler).Methods(http.MethodDelete)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	a.server = server
	a.store = db.NewMapDB()

	return nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

type RedirectResponse struct {
	OriginalURL string `json:"original_url"`
	Err         string `json:"error"`
}

// RedirectHandler handles GET access to shortened urls, this endpoint is publically available
// GET http://localhost:8080/foo
// TODO: respond with correct status codes
func (a *App) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	resp := &RedirectResponse{}
	vars := mux.Vars(r)
	w.Header().Set(ContentType, JSONMimeType)

	shortenedURL := vars["url"]
	if shortenedURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(resp)
		w.Write(b)
	}
	fmt.Printf("Handling request for shortened url %s\n", shortenedURL)

	storedURL, err := a.store.Get(shortenedURL)
	if err != nil {
		/// the error could possibly be more than not found, could be some db error, TODO:
		// distinguish between the two
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
	OriginalURL string `json:"original_url"`
}

type CreateResponse struct {
	ShortenedURL string `json:"shortened_url"`
	Err          string `json:"error"`
}

// CreateHandler handles the creation of new shortened URLS. TODO: This endpoint should be protected
// from public access
// curl localhost:8080/create -d '{"original_url": "http://foobarcat.blogspot.com"}'
// TODO: Add test case where mandatory field original_url is missing
func (a *App) CreateHandler(w http.ResponseWriter, r *http.Request) {
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

	shortenedURL, err := a.Create(req)

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

func (a *App) Create(req *CreateRequest) (string, error) {
	shortenedURL := Hash(req.OriginalURL)
	fmt.Printf("Create request for [%s], hashes to [%s]\n", req.OriginalURL, shortenedURL)
	// TODO: Check key isn't present, if it is we want to check if it is the same original URL
	// if it isn't then we need to resolve the collision, we could append some value to the original
	// and rehash

	storedURL := &db.StoredURL{req.OriginalURL}
	err := a.store.Create(shortenedURL, storedURL)
	return shortenedURL, err
}

func Hash(in string) string {
	hash := md5.Sum([]byte(in))
	s := base64.URLEncoding.EncodeToString((hash[:]))
	return s[0:LenShortened]
}

// If we add the concept of users do we want only the creator of a url to be able to delete it?
// If not, we could have some malicious user delete all the urls, leave unimplemented till decided
func (a *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}
