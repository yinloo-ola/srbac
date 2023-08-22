package store

import (
	"errors"
)

// Store is a generic interface to create, insert, update, retrieve, delete O.
// Note that O is a struct that might contain an array of primitive values or even structs
type Store[O any] interface {
	Insert(obj O) (int64, error)
	Update(id int64, obj O) error
	GetMulti(ids []int64) ([]O, error)
	GetOne(id int64) (O, error)
	GetAll() ([]O, error)
	FindField(field string, val any) ([]O, error)
	DeleteMulti(ids []int64) error
}

var ErrNotFound error = errors.New("record not found")
