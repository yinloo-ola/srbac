package store

import (
	"errors"
)

type RowScanner interface {
	Scan(dest ...any) error
}

// Row is a type constraint for types representing
// a single database row.
type Row[T any] interface {
	// FieldsVals returns all fields of a struct for use with row.Scan.
	FieldsVals() []any
	ScanRow(row RowScanner) error
	*T
}

// Store is a generic interface to create, insert, update, retrieve, delete O.
// Note that O is a struct that might contain an array of primitive values or even structs
type Store[T any, R Row[T]] interface {
	Insert(obj T) (int64, error)
	Update(id int64, obj T) error
	GetMulti(ids []int64) ([]T, error)
	GetOne(id int64) (T, error)
	GetAll() ([]T, error)
	FindField(field string, val any) ([]T, error)
	DeleteMulti(ids []int64) error
	Close() error
}

var ErrNotFound error = errors.New("record not found")
