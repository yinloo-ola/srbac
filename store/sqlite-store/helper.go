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
	IsIdxAsc   bool
	IsIdxDesc  bool
	IsIdxUniq  bool
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

func generateCreateIdxSQL(tableName string, columns []column) string {
	queries := make([]string, 0, len(columns))
	for _, col := range columns {
		uniq := ""
		if col.IsIdxUniq {
			uniq = "UNIQUE "
		}
		if col.IsIdxAsc {
			s := fmt.Sprintf("CREATE %sINDEX idx_%s ON %s (%s asc);", uniq, col.Name, tableName, col.Name)
			queries = append(queries, s)
		} else if col.IsIdxDesc {
			s := fmt.Sprintf("CREATE %sINDEX idx_%s ON %s (%s desc);", uniq, col.Name, tableName, col.Name)
			queries = append(queries, s)
		}
	}
	return strings.Join(queries, " ")
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

		isIdxAsc := false
		isIdxDesc := false
		if strings.Contains(tag, ",idx_asc") {
			isIdxAsc = true
		} else if strings.Contains(tag, ",idx_desc") {
			isIdxDesc = true
		}

		isUniqIdx := false
		if strings.Contains(tag, ",uniq") {
			isUniqIdx = true
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
			IsIdxAsc:   isIdxAsc,
			IsIdxDesc:  isIdxDesc,
			IsIdxUniq:  isUniqIdx,
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
