package shortly

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestReqUnmarshal(t *testing.T) {
	a := assert.New(t)
	r := &CreateRequest{}
	b := []byte(`{"original_url": "http://foobarcat.blogspot.com"}`)
	err := json.Unmarshal(b, r)
	spew.Dump(err)
	a.NoError(err)
}
