package sqlitestore

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/yinloo-ola/srbac/store"
)

type SQliteStore[T any, R store.Row[T]] struct {
	db         *sql.DB
	tablename  string
	pk         string
	getOneStmt *sql.Stmt
	insertStmt *sql.Stmt
	updateStmt *sql.Stmt
	getAllStmt *sql.Stmt
	columns    []column
	sync.RWMutex
}

func NewStore[T any, R store.Row[T]](path string) (*SQliteStore[T, R], error) {
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

	var obj T
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

	return &SQliteStore[T, R]{
		db: db, tablename: tableName, columns: columns, pk: pk,
		getOneStmt: getOneStmt, insertStmt: insertStmt, updateStmt: updateStmt,
		getAllStmt: getAllstmt,
	}, nil
}

func (o *SQliteStore[T, R]) Insert(obj T) (int64, error) {
	o.Lock()
	defer o.Unlock()
	values := make([]any, 0, len(o.columns))
	k := R(&obj)

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

func (o *SQliteStore[T, R]) Update(id int64, obj T) error {
	o.Lock()
	defer o.Unlock()
	values := make([]any, 0, len(o.columns))
	k := R(&obj)
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

func (o *SQliteStore[T, R]) GetMulti(ids []int64) ([]T, error) {
	o.RLock()
	defer o.RUnlock()
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

	objs := make([]T, 0, len(ids))
	for rows.Next() {
		var obj T
		k := R(&obj)
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

func (o *SQliteStore[T, R]) GetOne(id int64) (T, error) {
	o.RLock()
	defer o.RUnlock()
	var obj T
	k := R(&obj)

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

func (o *SQliteStore[T, R]) DeleteMulti(ids []int64) error {
	o.Lock()
	defer o.Unlock()
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

func (o *SQliteStore[T, R]) FindWhere(conds ...store.Cond) ([]T, error) {
	o.RLock()
	defer o.RUnlock()
	columnNames := make([]string, 0, len(o.columns))
	for _, col := range o.columns {
		columnNames = append(columnNames, col.Name)
	}
	whereStmt := ""
	stmts := make([]string, 0, len(conds))
	args := make([]any, 0, len(conds))
	for _, cond := range conds {
		s, arg := cond.GetQueryWithArgs()
		stmts = append(stmts, s)
		args = append(args, arg...)
	}
	if len(stmts) > 0 {
		whereStmt = " where " + strings.Join(stmts, " ")
	}
	findQuery := fmt.Sprintf("SELECT %s from %s%s", strings.Join(columnNames, ","), o.tablename, whereStmt)
	rows, err := o.db.Query(findQuery, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("%s GetMulti Query error: %w", o.tablename, err)
	}
	defer rows.Close()

	var objs []T
	for rows.Next() {
		var obj T
		k := R(&obj)
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

func (o *SQliteStore[T, R]) Close() error {
	return o.db.Close()
}
