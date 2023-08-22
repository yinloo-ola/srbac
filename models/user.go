package models

import (
	"encoding/json"

	"github.com/yinloo-ola/srbac/helper"
	"github.com/yinloo-ola/srbac/store"
)

type User struct {
	Id     int64   `db:"id,pk"`
	UserID string  `db:"user_id,idx_asc,uniq"`
	Roles  []int64 `db:"roles,json"`
}

func (o *User) FieldsVals() []any {
	roles, err := json.Marshal(o.Roles)
	helper.PanicErr(err)
	return []any{o.Id, o.UserID, roles}
}

func (o *User) ScanRow(row store.RowScanner) error {
	var roles []byte
	err := row.Scan(&o.Id, &o.UserID, &roles)
	if err != nil {
		return err
	}
	err = json.Unmarshal(roles, &o.Roles)
	helper.PanicErr(err)
	return nil
}
