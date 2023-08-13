package sqlitestore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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
	value := reflect.ValueOf(k)
	typ := value.Type()
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
	_, err := o.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}
func (o *SqliteStore[K]) GetMulti(ids []int64) ([]K, error) {
	return nil, errors.ErrUnsupported
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
	for rows.Next() {
		values := make([]interface{}, len(o.columns))
		valuePtrs := make([]interface{}, len(o.columns))
		for i := range o.columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return obj, fmt.Errorf("fail to scan values: %w", err)
		}

		objValue := reflect.ValueOf(&obj)
		elem := objValue.Elem()
		for _, col := range o.columns {
			field := elem.Field(col.Index)
			if field.IsValid() && field.CanSet() {
				v := values[col.Index]
				if col.IsJSON {
					var v2 any
					err = json.Unmarshal([]byte(v.(string)), &v2)
					if err != nil {
						return obj, fmt.Errorf("fail to unmarshal json text field: %w", err)
					}
					v = v2
				}
				if reflect.ValueOf(v).Type().AssignableTo(field.Type()) {
					field.Set(reflect.ValueOf(v))
				} else if reflect.ValueOf(v).Kind() == reflect.Slice && field.Kind() == reflect.Slice {
					sliceValue := reflect.MakeSlice(field.Type(), len(v.([]any)), len(v.([]any)))
					for i, val := range v.([]any) {
						elementValue := reflect.ValueOf(val)
						if !elementValue.Type().ConvertibleTo(field.Type().Elem()) {
							return obj, fmt.Errorf("fail to populate slice. Wrong type: %s", elementValue.Type())
						}
						sliceValue.Index(i).Set(elementValue.Convert(field.Type().Elem()))
					}
					field.Set(sliceValue)
				}
			}
		}
	}
	return obj, nil
}
func (o *SqliteStore[K]) GetAll() ([]K, error) {
	return nil, errors.ErrUnsupported
}
func (o *SqliteStore[K]) DeleteMulti(ids []int64) error {
	return errors.ErrUnsupported
}