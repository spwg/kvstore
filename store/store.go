// Package store provides functionality for a key-value store.
package store

import "context"

// Store defines an interface that all implements of a key-value store will adhere to.
type Store interface {
	// Create creates a new entry in the database.
	//
	// key will become the key of the entry, and it can be used it Read, Update, and Delete.
	// It must be unique.
	//
	// value will become mapped to by key.
	Create(ctx context.Context, key, value []byte) error

	// Read returns the value mapped to from key, which must exist already.
	Read(ctx context.Context, key []byte) ([]byte, error)

	// Update edits the value that key points to with another value.
	Update(ctx context.Context, key, value []byte) error

	// Delete removes the entry identified by key from the database, which must exist.
	Delete(ctx context.Context, key []byte) error
}
