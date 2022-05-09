// Package memstore provides functionality for an in-memory key-value store.
package memstore

import (
	"context"
	"fmt"
	"hash/maphash"
	"kvstore/store"
	"sync"
)

// Store defines a key-value store implementing store.Store.
type Store struct {
	kv   map[uint64][]byte
	hash *maphash.Hash
	mu   *sync.Mutex
}

var _ store.Store = (*Store)(nil)

func (s *Store) Create(ctx context.Context, key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(key) == 0 {
		return fmt.Errorf("invalid key: cannot have zero-length")
	}
	hash := s.hashKey(key)
	if _, ok := s.kv[hash]; ok {
		return fmt.Errorf("key already exists")
	}
	s.kv[hash] = value
	return nil
}

func (s *Store) Read(ctx context.Context, key []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := s.hashKey(key)
	v, ok := s.kv[hash]
	if !ok {
		return nil, fmt.Errorf("key does not exist")
	}
	return v, nil
}

func (s *Store) Update(ctx context.Context, key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := s.hashKey(key)
	if _, ok := s.kv[hash]; !ok {
		return fmt.Errorf("key does not exist")
	}
	s.kv[hash] = value
	return nil
}

func (s *Store) Delete(ctx context.Context, key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := s.hashKey(key)
	if _, ok := s.kv[hash]; !ok {
		return fmt.Errorf("key does not exist")
	}
	delete(s.kv, hash)
	return nil
}

// hashKey returns a hash for the given key. Must be called with s.mu locked.
func (s *Store) hashKey(key []byte) uint64 {
	s.hash.Reset()
	_, _ = s.hash.Write(key) // never fails
	return s.hash.Sum64()
}

// New initializes a *Store.
func New() *Store {
	return &Store{
		kv:   make(map[uint64][]byte),
		hash: &maphash.Hash{},
		mu:   &sync.Mutex{},
	}
}
