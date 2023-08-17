package sqlitestore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/yinloo-ola/srbac/store"
	_ "modernc.org/sqlite"
)

type SqliteStore[K any] struct {
	db        *sql.DB
	tablename string
	pk        string
	columns   []column
}

func NewStore[K any](path string) (*SqliteStore[K], error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	var k K
	typ := reflect.TypeOf(k)
	tableName := toSnakeCase(typ.Name())
	columns := getColumns(typ)

	pk := ""
	for _, col := range columns {
		if col.IsPK {
			pk = col.Name
			break
		}
	}

	sql := generateCreateTableSQL(tableName, columns)
	_, err = db.Exec(sql)
	if err != nil {
		return nil, err
	}

	return &SqliteStore[K]{db: db, tablename: tableName, columns: columns, pk: pk}, nil
}

func (o *SqliteStore[K]) Insert(obj K) (int64, error) {
	value := reflect.ValueOf(obj)

	columns := make([]string, 0, len(o.columns))
	placeholders := make([]string, 0, len(o.columns))
	values := make([]any, 0, len(o.columns))
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		columns = append(columns, col.Name)
		placeholders = append(placeholders, "?")
		val := value.Field(col.Index).Interface()
		if col.IsJSON {
			s, err := json.Marshal(val)
			if err != nil {
				return 0, fmt.Errorf("fail to json marshal json field: %w", err)
			}
			val = string(s)
		}
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
func (o *SqliteStore[K]) Update(id int64, obj K) error {
	value := reflect.ValueOf(obj)

	columns := make([]string, 0, len(o.columns))
	values := make([]any, 0, len(o.columns))
	for _, col := range o.columns {
		if col.IsPK {
			continue
		}
		columns = append(columns, col.Name+"=?")
		val := value.Field(col.Index).Interface()
		if col.IsJSON {
			s, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("fail to json marshal json field: %w", err)
			}
			val = string(s)
		}
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
func (o *SqliteStore[K]) GetMulti(ids []int64) ([]K, error) {
	return nil, errors.New("not implemented yet")
}
func (o *SqliteStore[K]) GetOne(id int64) (K, error) {
	var obj K
	columns := make([]string, 0, len(o.columns))
	for _, col := range o.columns {
		columns = append(columns, col.Name)
	}
	query := fmt.Sprintf("SELECT %s from %s where %s=?", strings.Join(columns, ","), o.tablename, o.pk)
	rows, err := o.db.Query(query, id)
	if err != nil {
		return obj, fmt.Errorf("select failed: %w", err)
	}
	defer rows.Close()
	rowCount := 0
	for rows.Next() {
		rowCount++
		values := make([]interface{}, len(o.columns))
		valuePtrs := make([]interface{}, len(o.columns))
		for i := range o.columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return obj, fmt.Errorf("fail to scan values: %w", err)
		}

		now := time.Now()
		objValue := reflect.ValueOf(&obj)
		fmt.Println("duration [reflect.ValueOf]:", time.Since(now))
		elem := objValue.Elem()
		for _, col := range o.columns {
			now = time.Now()
			field := elem.Field(col.Index)
			fmt.Println("duration [elem.Field]:", time.Since(now))
			if field.IsValid() && field.CanSet() {
				v := values[col.Index]
				if col.IsJSON {
					err = unmarshalJSONIntoValue(v.(string), field)
					if err != nil {
						return obj, fmt.Errorf("fail to unmarshal json into field: %w", err)
					}
				} else {
					if reflect.TypeOf(v).AssignableTo(field.Type()) {
						now = time.Now()
						field.Set(reflect.ValueOf(v))
						fmt.Println("duration [field.Set non bool]:", col.SqLiteType, time.Since(now))
					} else if col.IsBool {
						now = time.Now()
						field.SetBool(v.(int64) > 0)
						fmt.Println("duration [field.Set bool]:", col.SqLiteType, time.Since(now))
					} else {
						return obj, fmt.Errorf("fail to set value")
					}
				}
			}
		}
	}
	if rowCount == 0 {
		return obj, store.ErrNotFound
	}
	return obj, nil
}
func (o *SqliteStore[K]) GetAll() ([]K, error) {
	return nil, errors.New("not implemented yet")
}
func (o *SqliteStore[K]) DeleteMulti(ids []int64) error {
	return errors.New("not implemented yet")
}
