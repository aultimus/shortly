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

type MapDB struct {
	m map[string]*StoredURL
}

func (m *MapDB) Create(key string, value *StoredURL) error {
	// TODO: handle collsion - is this done per user?
	m.m[key] = value
	return nil
}

func (m *MapDB) Get(key string) (*StoredURL, error) {
	var err error
	value, exists := m.m[key]
	if !exists {
		err = fmt.Errorf("key %s does not exist in db", key)
	}
	return value, err
}

func NewMapDB() *MapDB {
	return &MapDB{
		m: make(map[string]*StoredURL),
	}
}
