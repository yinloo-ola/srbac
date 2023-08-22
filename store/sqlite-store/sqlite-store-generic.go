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

type SQliteStore[O any, R Row[O]] struct {
	db         *sql.DB
	tablename  string
	pk         string
	getOneStmt *sql.Stmt
	insertStmt *sql.Stmt
	updateStmt *sql.Stmt
	getAllStmt *sql.Stmt
	columns    []column
}

func NewStore[O any, K Row[O]](path string) (*SQliteStore[O, K], error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA journal_mode = wal;")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA synchronous=1;")
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

	placeholdersNoPK := make([]string, 0, len(columns))
	columnNames := make([]string, 0, len(columns))
	columnNamesNoPK := make([]string, 0, len(columns))
	updates := make([]string, 0, len(columns))
	for _, col := range columns {
		columnNames = append(columnNames, col.Name)
		if !col.IsPK {
			columnNamesNoPK = append(columnNamesNoPK, col.Name)
			placeholdersNoPK = append(placeholdersNoPK, "?")
			updates = append(updates, col.Name+"=?")
		}
	}

	getOneQuery := fmt.Sprintf("SELECT %s from %s where %s=?", strings.Join(columnNames, ","), tableName, pk)
	getOneStmt, err := db.Prepare(getOneQuery)
	if err != nil {
		return nil, err
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columnNamesNoPK, ", "),
		strings.Join(placeholdersNoPK, ", "),
	)
	insertStmt, err := db.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}

	updateQuery := fmt.Sprintf("UPDATE %s SET %s where %s=?",
		tableName,
		strings.Join(updates, ", "),
		pk,
	)
	updateStmt, err := db.Prepare(updateQuery)
	if err != nil {
		return nil, err
	}

	getAllQuery := fmt.Sprintf("SELECT %s from %s", strings.Join(columnNames, ","), tableName)
	getAllstmt, err := db.Prepare(getAllQuery)
	if err != nil {
		return nil, err
	}

	return &SQliteStore[O, K]{
		db: db, tablename: tableName, columns: columns, pk: pk,
		getOneStmt: getOneStmt, insertStmt: insertStmt, updateStmt: updateStmt,
		getAllStmt: getAllstmt,
	}, nil
}

func (o *SQliteStore[O, K]) Insert(obj O) (int64, error) {
	values := make([]any, 0, len(o.columns))
	k := K(&obj)

	fieldPtrs := k.FieldsVals()
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		val := fieldPtrs[col.Index]
		values = append(values, val)
	}

	res, err := o.insertStmt.Exec(values...)
	if err != nil {
		return 0, fmt.Errorf("%s insert failed: %w", o.tablename, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s fail to get last insert id: %w", o.tablename, err)
	}
	return id, nil
}

func (o *SQliteStore[O, K]) Update(id int64, obj O) error {
	values := make([]any, 0, len(o.columns))
	k := K(&obj)
	fieldPtrs := k.FieldsVals()
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		val := fieldPtrs[col.Index]
		values = append(values, val)
	}
	values = append(values, id)

	res, err := o.updateStmt.Exec(values...)
	if err != nil {
		return fmt.Errorf("%s update failed: %w", o.tablename, err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s failed to get rows affected: %w", o.tablename, err)
	} else if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (o *SQliteStore[O, K]) GetMulti(ids []int64) ([]O, error) {
	columnNames := make([]string, 0, len(o.columns))
	for _, col := range o.columns {
		columnNames = append(columnNames, col.Name)
	}

	placeholders, args := InArgs(ids)
	query := fmt.Sprintf("SELECT %s from %s where %s in (%s)", strings.Join(columnNames, ","), o.tablename, o.pk, placeholders)

	rows, err := o.db.Query(query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("%s GetMulti Query error: %w", o.tablename, err)
	}
	defer rows.Close()

	objs := make([]O, 0, len(ids))
	for rows.Next() {
		var obj O
		k := K(&obj)
		err = k.ScanRow(rows)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, store.ErrNotFound
			}
			return nil, fmt.Errorf("%s GetMulti row.Scan error: %w", o.tablename, err)
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (o *SQliteStore[O, K]) GetOne(id int64) (O, error) {
	var obj O
	k := K(&obj)

	row := o.getOneStmt.QueryRow(id)
	if row == nil {
		return obj, store.ErrNotFound
	}

	err := k.ScanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return obj, store.ErrNotFound
		}
		return obj, fmt.Errorf("%s GetOne row.Scan error: %w", o.tablename, err)
	}
	return obj, nil
}
func (o *SQliteStore[O, K]) GetAll() ([]O, error) {
	rows, err := o.getAllStmt.Query()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("%s GetMulti Query error: %w", o.tablename, err)
	}
	defer rows.Close()

	objs := make([]O, 0, 100)
	for rows.Next() {
		var obj O
		k := K(&obj)
		err = k.ScanRow(rows)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, store.ErrNotFound
			}
			return nil, fmt.Errorf("%s GetMulti row.Scan error: %w", o.tablename, err)
		}
		objs = append(objs, obj)
	}
	return objs, nil
}
func (o *SQliteStore[O, K]) DeleteMulti(ids []int64) error {
	placeholder, args := InArgs(ids)
	query := fmt.Sprintf("DELETE from %s where %s IN (%s)", o.tablename, o.pk, placeholder)
	res, err := o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s DeleteMulti exec failed: %w", o.tablename, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s DeleteMulti RowsAffected failed: %w", o.tablename, err)
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}
func (o *SQliteStore[O, K]) FindField(field string, val any) ([]O, error) {
	columnNames := make([]string, 0, len(o.columns))
	for _, col := range o.columns {
		columnNames = append(columnNames, col.Name)
	}
	findQuery := fmt.Sprintf("SELECT %s from %s where %s=?", strings.Join(columnNames, ","), o.tablename, field)
	rows, err := o.db.Query(findQuery, val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("%s GetMulti Query error: %w", o.tablename, err)
	}
	defer rows.Close()

	var objs []O
	for rows.Next() {
		var obj O
		k := K(&obj)
		err = k.ScanRow(rows)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, store.ErrNotFound
			}
			return nil, fmt.Errorf("%s GetMulti row.Scan error: %w", o.tablename, err)
		}
		objs = append(objs, obj)
	}
	return objs, nil
}
