package models

import (
	"encoding/json"

	"github.com/yinloo-ola/srbac/helper"
	sqlitestore "github.com/yinloo-ola/srbac/store/sqlite-store"
)

type Role struct {
	Id          int64   `db:"id,pk"`
	Name        string  `db:"name"`
	Description string  `db:"description"`
	Permissions []int64 `db:"permissions,json"`
}

func (o *Role) FieldsVals() []any {
	perms, err := json.Marshal(o.Permissions)
	helper.PanicErr(err)
	return []any{o.Id, o.Name, o.Description, perms}
}

func (o *Role) ScanRow(row sqlitestore.RowScanner) error {
	var perms []byte
	err := row.Scan(&o.Id, &o.Name, &o.Description, &perms)
	if err != nil {
		return err
	}
	err = json.Unmarshal(perms, &o.Permissions)
	helper.PanicErr(err)
	return nil
}
