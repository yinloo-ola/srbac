package sqlitestore

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/yinloo-ola/srbac/store"
)

type SQliteStore[O any, R store.Row[O]] struct {
	db         *sql.DB
	tablename  string
	pk         string
	getOneStmt *sql.Stmt
	columns    []column
}

func NewStore[O any, K store.Row[O]](path string) (*SQliteStore[O, K], error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA journal_mode = wal;")
	if err != nil {
		return nil, err
	}

	var obj O
	typ := reflect.TypeOf(obj)
	tableName := toSnakeCase(typ.Name())
	columns := getColumns(typ)

	pk := ""
	for _, col := range columns {
		if col.IsPK {
			pk = col.Name
			break
		}
	}

	stmt := generateCreateTableSQL(tableName, columns)
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, err
	}

	stmt = generateCreateIdxSQL(tableName, columns)
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, err
	}

	columnNames := make([]string, 0, len(columns))
	for _, col := range columns {
		columnNames = append(columnNames, col.Name)
	}
	query := fmt.Sprintf("SELECT %s from %s where %s=?", strings.Join(columnNames, ","), tableName, pk)
	getOneStmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}

	return &SQliteStore[O, K]{db: db, tablename: tableName, columns: columns, pk: pk, getOneStmt: getOneStmt}, nil
}

func (o *SQliteStore[O, K]) Insert(obj O) (int64, error) {
	columns := make([]string, 0, len(o.columns))
	placeholders := make([]string, 0, len(o.columns))
	values := make([]any, 0, len(o.columns))
	k := K(&obj)

	fieldPtrs := k.FieldsVals()
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		columns = append(columns, col.Name)
		placeholders = append(placeholders, "?")
		val := fieldPtrs[col.Index]
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		o.tablename,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))
	res, err := o.db.Exec(query, values...)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("fail to get last insert id: %w", err)
	}
	return id, nil
}

func (o *SQliteStore[O, K]) Update(id int64, obj O) error {
	columns := make([]string, 0, len(o.columns))
	values := make([]any, 0, len(o.columns))
	k := K(&obj)
	fieldPtrs := k.FieldsVals()
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		columns = append(columns, col.Name+"=?")
		val := fieldPtrs[col.Index]
		values = append(values, val)
	}
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s where %s=?",
		o.tablename,
		strings.Join(columns, ", "),
		o.pk,
	)
	res, err := o.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (o *SQliteStore[O, K]) GetMulti(ids []int64) ([]K, error) {
	return nil, errors.New("not implemented yet")
}

func (o *SQliteStore[O, K]) GetOne(id int64) (O, error) {
	var obj O
	k := K(&obj)

	row := o.getOneStmt.QueryRow(id)
	err := k.Scan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return obj, store.ErrNotFound
		}
		return obj, fmt.Errorf("SQliteStore %s GetOne row.Scan error: %w", o.tablename, err)
	}
	return obj, nil
}
func (o *SQliteStore[O, K]) GetAll() ([]K, error) {
	return nil, errors.New("not implemented yet")
}
func (o *SQliteStore[O, K]) DeleteMulti(ids []int64) error {
	return errors.New("not implemented yet")
}
