package models

import sqlitestore "github.com/yinloo-ola/srbac/store/sqlite-store"

type Permission struct {
	Id          int64  `db:"id,pk"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (o *Permission) FieldsVals() []any {
	return []any{o.Id, o.Name, o.Description}
}

func (o *Permission) ScanRow(row sqlitestore.RowScanner) error {
	return row.Scan(&o.Id, &o.Name, &o.Description)
}
