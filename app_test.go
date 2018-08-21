package shortly

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aultimus/shortly/db"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

type DBErrStore struct {
}

func (s *DBErrStore) Create(string, *db.StoredURL) error {
	return db.NewErrDB("foo")
}

func (s *DBErrStore) Get(string) (*db.StoredURL, error) {
	return nil, db.NewErrDB("foo")
}

func TestRedirectHandler(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	app.Init(db.NewMapDB())
	req, err := http.NewRequest("GET", "/redirect/foo", nil)
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
	stored := &db.StoredURL{OriginalURL: "http://redirect/foofoo.com/bar"}
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

	// force DBError and check we return internal server error
	app.store = &DBErrStore{}
	req, err = http.NewRequest("GET", "/redirect/cat", nil)
	a.NoError(err)

	rr = httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	a.Equal(http.StatusInternalServerError, rr.Code)
}

func TestCreateHandler(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	app.Init(db.NewMapDB())
	body := strings.NewReader(`{"original_url": "http://foobarcat.blogspot.com"}`)
	req, err := http.NewRequest("POST", "/create", body)
	a.NoError(err)

	// create new entry
	rr := httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	// check results
	a.Equal(http.StatusOK, rr.Code)
	resp := &CreateResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp)
	a.NoError(err)
	a.Empty(resp.Err)
	a.NotEmpty(resp.ShortenedURL)

	// create same url, check that we get the same short url back
	body = strings.NewReader(`{"original_url": "http://foobarcat.blogspot.com"}`)
	req, err = http.NewRequest("POST", "/create", body)
	rr = httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	// check results
	a.Equal(http.StatusOK, rr.Code)
	resp1 := &CreateResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp1)
	a.NoError(err)
	a.Empty(resp1.Err)
	a.NotEmpty(resp1.ShortenedURL)
	a.Equal(resp.ShortenedURL, resp1.ShortenedURL)

	// create different url, check we get different short url back
	body = strings.NewReader(`{"original_url": "http://www.google.com"}`)
	req, err = http.NewRequest("POST", "/create", body)
	rr = httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	// check results
	a.Equal(http.StatusOK, rr.Code)
	resp2 := &CreateResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp2)
	a.NoError(err)
	a.Empty(resp2.Err)
	a.NotEmpty(resp2.ShortenedURL)
	a.NotEqual(resp2.ShortenedURL, resp.ShortenedURL)

	// force DBError and check we return internal server error
	app.store = &DBErrStore{}
	body = strings.NewReader(`{"original_url": "http://www.google.com"}`)
	req, err = http.NewRequest("POST", "/create", body)
	rr = httptest.NewRecorder()
	app.server.Handler.ServeHTTP(rr, req) // kind of hacky

	// check results
	a.Equal(http.StatusInternalServerError, rr.Code)
	resp3 := &CreateResponse{}
	err = json.Unmarshal([]byte(rr.Body.String()), resp3)
	a.Empty(resp3.ShortenedURL)

	// TODO: create collision - check that we got different URL back, easily done when we enable
	// the custom_alias feature
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

	hasher := MD5Hash{}

	shortURL := hasher.Hash(longURL)
	shortURL1 := hasher.Hash(longURL)
	a.Equal(shortURL, shortURL1, "check the same long url generates the same short url")
	a.Equal(LenShortened, len(shortURL), "check that the length of the short url is as desired")

	shortURL2 := hasher.Hash(longURL2)
	a.NotEqual(shortURL, shortURL2, "check that different long urls generate different short urls")

	shortURL3 := hasher.Hash(longURL[0 : len(longURL)-2])
	a.NotEqual(longURL, shortURL3, "check that a subset string doesnt generate same long url")

	shortURL4 := hasher.Hash(longURL + "f")
	a.NotEqual(longURL, shortURL4, "check that a superset string doesnt generate same long url")
}

type Collision struct {
	maxCollisions int
	numCollisions int
}

func (c *Collision) Hash(in string) string {
	if c.numCollisions > c.maxCollisions {
		return "bar"
	}
	c.numCollisions++
	return "foo"
}

func TestCollision(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	dbMap := db.NewMapDB()
	dbMap.M["foo"] = &db.StoredURL{"bar"}
	app.Init(dbMap)

	_, err := app.Create(&CreateRequest{"http://www.google.com"}, &Collision{maxCollisions: 64})
	a.Error(err)

	_, ok := err.(*db.ErrCollision)
	a.True(ok)
}

func TestSomeCollision(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	dbMap := db.NewMapDB()
	dbMap.M["foo"] = &db.StoredURL{"bar"}
	app.Init(dbMap)

	_, err := app.Create(&CreateRequest{"http://www.google.com"}, &Collision{maxCollisions: 63})
	a.NoError(err)
}

func TestCreateSameData(t *testing.T) {
	a := assert.New(t)

	app := NewApp()
	dbMap := db.NewMapDB()
	dbMap.M["foo"] = &db.StoredURL{"bar"}
	app.Init(dbMap)

	_, err := app.Create(&CreateRequest{"bar"}, &MD5Hash{})
	a.NoError(err)
}

// TODO: write more Hash collision errors such as one that only succeeds after x tries
// also check that the stored URL / key is alright
