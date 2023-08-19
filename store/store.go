package store

import (
	"database/sql"
	"errors"
)

// Row is a type constraint for types representing
// a single database row.
type Row[T any] interface {
	// PtrFields returns all fields of a struct for use with row.Scan.
	// It must be implemented with a pointer receiver type, and all elements
	// in the returned slice must also be pointers.
	FieldsVals() []any
	Scan(row *sql.Row) error
	*T
}

// Store is a generic interface to create, insert, update, retrieve, delete O.
// Note that O is a struct that might contain an array of primitive values or even structs
type Store[O any, K Row[O]] interface {
	Insert(obj O) (int64, error)
	Update(id int64, obj O) error
	GetMulti(ids []int64) ([]O, error)
	GetOne(id int64) (O, error)
	GetAll() ([]O, error)
	DeleteMulti(ids []int64) error
}

var ErrNotFound error = errors.New("record not found")
