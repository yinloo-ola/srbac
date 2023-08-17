package sqlitestore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

type column struct {
	Name       string
	Index      int
	IsPK       bool
	IsJSON     bool
	SqLiteType sqliteType
}
type sqliteType string

const (
	sqliteTypeText sqliteType = "TEXT"
	sqliteTypeInt  sqliteType = "INTEGER"
	sqliteTypeReal sqliteType = "REAL"
)

func generateCreateTableSQL(tableName string, columns []column) string {
	return fmt.Sprintf("CREATE TABLE if not exists %s (%s)", tableName, generateCreateColumnSQL(columns))
}

func generateCreateColumnSQL(columns []column) string {
	colStrings := make([]string, 0, len(columns))
	for _, col := range columns {
		s := fmt.Sprintf("%s %s", col.Name, col.SqLiteType)
		if col.IsPK {
			s += " PRIMARY KEY"
		}
		colStrings = append(colStrings, s)
	}
	return strings.Join(colStrings, ", ")
}

func getColumns(typ reflect.Type) []column {
	var columns []column

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")

		isPK := false
		if strings.Contains(tag, ",pk") {
			isPK = true
		}

		IsJSON := false
		if strings.Contains(tag, ",json") {
			IsJSON = true
		}

		name := field.Name
		splitted := strings.Split(tag, ",")
		if len(splitted) > 0 && len(splitted[0]) > 0 {
			name = splitted[0]
		}

		sqlType := getSQLiteType(field.Type)

		columns = append(columns, column{
			Name:       name,
			Index:      i,
			IsPK:       isPK,
			IsJSON:     IsJSON,
			SqLiteType: sqlType,
		})
	}
	return columns
}

func getSQLiteType(field reflect.Type) sqliteType {
	switch field.Kind() {
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Int16, reflect.Int32, reflect.Int8:
		return sqliteTypeInt
	case reflect.Bool:
		return sqliteTypeInt
	case reflect.String:
		return sqliteTypeText
	case reflect.Float32, reflect.Float64:
		return sqliteTypeReal
	case reflect.Struct:
		return sqliteTypeText
	case reflect.Pointer:
		if isPrimitive(field.Elem().Kind()) {
			panic("pointer to primitive is not supported")
		}
		return sqliteTypeText
	case reflect.Array:
		return sqliteTypeText
	case reflect.Slice:
		return sqliteTypeText
	default:
		panic("unsupported type")
	}
}

func isPrimitive(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

func toSnakeCase(input string) string {
	var result []rune

	for i, char := range input {
		if i > 0 && (unicode.IsUpper(char) || unicode.IsDigit(char)) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(char))
	}

	return string(result)
}

func unmarshalJSONIntoValue(jsonStr string, value reflect.Value) error {
	if !value.CanAddr() {
		return fmt.Errorf("value must be addressable")
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}

		return json.Unmarshal([]byte(jsonStr), value.Interface())

	case reflect.Struct:
		return json.Unmarshal([]byte(jsonStr), value.Addr().Interface())

	case reflect.Slice:
		now := time.Now()
		elemKind := value.Type().Elem().Kind()
		fmt.Println("duration [value.Type().Elem().Kind()]:", time.Since(now))

		if elemKind == reflect.Ptr {
			now = time.Now()
			tempSlice := reflect.New(value.Type()).Elem()
			fmt.Println("duration [reflect.New]:", time.Since(now))
			err := json.Unmarshal([]byte(jsonStr), tempSlice.Addr().Interface())
			if err != nil {
				return err
			}

			now = time.Now()
			value.Set(tempSlice)
			fmt.Println("duration [value.Set]:", time.Since(now))

			return nil
		} else {
			return json.Unmarshal([]byte(jsonStr), value.Addr().Interface())
		}

	default:
		return fmt.Errorf("unsupported value kind: %s", value.Kind().String())
	}
}
func setReflectValue(obj any, colType sqliteType, value reflect.Value) error {
	switch colType {
	case sqliteTypeText:
		value.SetString(obj.(string))
	case sqliteTypeInt:
		value.SetInt(obj.(int64))
	case sqliteTypeReal:
		value.SetFloat(obj.(float64))
	}
	return fmt.Errorf("unsupported value kind: %s", colType)
}
