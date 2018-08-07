package shortly

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aultimus/shortly/db"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHandler(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	app.Init(db.NewMapDB())
	req, err := http.NewRequest("GET", "/foo", nil)
	a.NoError(err)

	// we miss the empty in an empty store
	rr := httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	// check results
	a.Equal(http.StatusNotFound, rr.Code)
	resp := &RedirectResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp)
	a.NoError(err)
	a.NotEmpty(resp.Err)
	a.Empty(resp.OriginalURL)

	// add entry
	stored := &db.StoredURL{OriginalURL: "http://foofoo.com/bar"}
	err = app.store.Create("foo", stored)
	a.NoError(err)

	// check we hit the entry
	rr = httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	a.Equal(http.StatusOK, rr.Code)
	// Check results
	resp = &RedirectResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp)
	a.NoError(err)
	a.Empty(resp.Err)
	a.Equal(stored.OriginalURL, resp.OriginalURL)
}

func TestReqUnmarshal(t *testing.T) {
	a := assert.New(t)
	r := &CreateRequest{}
	b := []byte(`{"original_url": "http://foobarcat.blogspot.com"}`)
	err := json.Unmarshal(b, r)
	spew.Dump(err)
	a.NoError(err)
}

func TestHash(t *testing.T) {
	a := assert.New(t)
	longURL := "http://foobarcat.blogspot.com"
	longURL2 := "http://www.google.com"

	shortURL := Hash(longURL)
	shortURL1 := Hash(longURL)
	a.Equal(shortURL, shortURL1, "check the same long url generates the same short url")
	a.Equal(LenShortened, len(shortURL), "check that the length of the short url is as desired")

	shortURL2 := Hash(longURL2)
	a.NotEqual(shortURL, shortURL2, "check that different long urls generate different short urls")

	shortURL3 := Hash(longURL[0 : len(longURL)-2])
	a.NotEqual(longURL, shortURL3, "check that a subset string doesnt generate same long url")

	shortURL4 := Hash(longURL + "f")
	a.NotEqual(longURL, shortURL4, "check that a superset string doesnt generate same long url")
}
