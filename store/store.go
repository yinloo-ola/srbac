package store

import "errors"

// Store is a generic interface to create, insert, update, retrieve, delete K.
// Note that K is a struct that might contain an array of primitive values or even structs
type Store[K any] interface {
	Insert(obj K) (int64, error)
	Update(id int64, obj K) error
	GetMulti(ids []int64) ([]K, error)
	GetOne(id int64) (K, error)
	GetAll() ([]K, error)
	DeleteMulti(ids []int64) error
}

var ErrNotFound error = errors.New("record not found")
