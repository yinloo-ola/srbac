package store

import (
	"errors"
	"fmt"
	"strings"
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

type WhereCond struct {
	Field string
	Op    op
	Val   any
}

type op string

const OpEqual op = "="
const OpNotEqual op = "<>"
const OpGte op = ">="
const OpGt op = ">"
const OpLte op = "<="
const OpLt op = "<"
const OpIn op = "in"

type QueryJoiner string

const QueryJoinerAnd QueryJoiner = "and"
const QueryJoinerOr QueryJoiner = "or"

func (o QueryJoiner) GetQueryWithArgs() (string, []any) {
	return string(o), []any{}
}

func (o WhereCond) GetQueryWithArgs() (string, []any) {
	switch o.Op {
	case OpIn:
		vals, ok := o.Val.([]any)
		if !ok {
			panic("WhereCond with OpIn only accept []any as Val")
		}
		qnMarks := make([]string, 0, len(vals))
		for range vals {
			qnMarks = append(qnMarks, "?")
		}
		return fmt.Sprintf("%s %s (%s)", o.Field, o.Op, strings.Join(qnMarks, ",")), vals
	default:
		return fmt.Sprintf("%s %s ?", o.Field, o.Op), []any{o.Val}
	}
}

type Cond interface {
	GetQueryWithArgs() (string, []any)
}

// Store is a generic interface to create, insert, update, retrieve, delete O.
// Note that O is a struct that might contain an array of primitive values or even structs
type Store[T any, R Row[T]] interface {
	Insert(obj T) (int64, error)
	Update(id int64, obj T) error
	GetMulti(ids []int64) ([]T, error)
	GetOne(id int64) (T, error)
	// FindWhere WhereConds must be either empty or joined by QueryJoiners
	FindWhere(...Cond) ([]T, error)
	DeleteMulti(ids []int64) error
	Close() error
}

var ErrNotFound error = errors.New("record not found")
