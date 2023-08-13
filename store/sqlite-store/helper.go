package sqlitestore

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type column struct {
	Name       string
	Index      int
	IsPK       bool
	IsJSON     bool
	SqLiteType string
}

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

func getSQLiteType(field reflect.Type) string {
	switch field.Kind() {
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Int16, reflect.Int32, reflect.Int8:
		return "INTEGER"
	case reflect.String:
		return "TEXT"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Struct:
		return "TEXT"
	case reflect.Array:
		return "TEXT"
	case reflect.Slice:
		return "TEXT"
	default:
		panic("unsupported type")
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
