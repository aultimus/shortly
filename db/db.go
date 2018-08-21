package db

import (
	"fmt"
)

type DBer interface {
	Create(string, *StoredURL) error
	Get(string) (*StoredURL, error)
}

type StoredURL struct {
	OriginalURL string `json:"original_string"`
}

type ErrBase struct {
	Message string
}

func (e *ErrBase) Error() string {
	return e.Message
}

type ErrCollision struct {
	ErrBase
}

func NewErrCollision(message string) *ErrCollision {
	return &ErrCollision{
		ErrBase: ErrBase{message},
	}
}

type ErrDB struct {
	ErrBase
}

func NewErrDB(message string) *ErrDB {
	return &ErrDB{
		ErrBase: ErrBase{message},
	}
}

type ErrNotFound struct {
	ErrBase
}

func NewErrNotFound(message string) error {
	return &ErrNotFound{
		ErrBase: ErrBase{message},
	}
}

type MapDB struct {
	M map[string]*StoredURL
}

func (m *MapDB) Create(key string, value *StoredURL) error {
	stored, exists := m.M[key]
	if exists {
		if *stored == *value {
			return nil
		}
	}
	m.M[key] = value
	return nil
}

func (m *MapDB) Get(key string) (*StoredURL, error) {
	var err error
	value, exists := m.M[key]
	if !exists {
		return nil, NewErrNotFound(fmt.Sprintf("key %s does not exist in db", key))
	}
	return value, err
}

func NewMapDB() *MapDB {
	return &MapDB{
		M: make(map[string]*StoredURL),
	}
}
