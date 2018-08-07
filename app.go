package shortly

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aultimus/shortly/db"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
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
	spew.Dump(string(b))
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
	shortenedURL := doHash(req.OriginalURL)

	storedURL := &db.StoredURL{req.OriginalURL}
	err := a.store.Create(shortenedURL, storedURL)
	return shortenedURL, err
}

func doHash(string) string {
	return "foo" // TODO
}

func (a *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}
