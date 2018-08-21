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

type ErrCollision struct {
	message string
}

func NewErrCollision(message string) *ErrCollision {
	return &ErrCollision{
		message: message,
	}
}

func (e *ErrCollision) Error() string {
	return e.message
}

type ErrDB struct {
	message string
}

func NewErrDB(message string) *ErrDB {
	return &ErrDB{
		message: message,
	}
}

func (e *ErrDB) Error() string {
	return e.message
}

// TODO:
// It's crazy to have all these lines to declare an error type, can we embed some basic behaviour?
type ErrNotFound struct {
	message string
}

func (e *ErrNotFound) Error() string {
	return e.message
}

func NewErrNotFound(message string) error {
	return &ErrNotFound{
		message: message,
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
		} else {
			return NewErrCollision(fmt.Sprintf("key %s already exists with a different value", key))
		}
	}
	m.M[key] = value
	return nil
}

func (m *MapDB) Get(key string) (*StoredURL, error) {
	var err error
	value, exists := m.M[key]
	if !exists {
		err = fmt.Errorf("key %s does not exist in db", key)
	}
	return value, err
}

func NewMapDB() *MapDB {
	return &MapDB{
		M: make(map[string]*StoredURL),
	}
}
