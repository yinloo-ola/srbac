package sqlitestore

import (
	"database/sql"
	"encoding/json"
)

type Role struct {
	Name         string     `db:"name,idx_desc,uniq"`
	IsHuman      bool       `db:"isHuman,idx_asc"`
	Permissions  []int64    `db:"permissions"`
	Ages         []int16    `db:"ages"`
	Alias        []string   `db:"alias"`
	Prices       []float32  `db:"prices"`
	Address      Address    `db:"address"`
	AddressPtr   *Address   `db:"addressPtr"`
	Addresses    []Address  `db:"addresses"`
	AddressesPtr []*Address `db:"addressesPtr"`
	Id           int64      `db:"id,pk"`
}

func (o *Role) FieldsVals() []any {
	permsStr, err := json.Marshal(o.Permissions)
	panicErr(err)

	agesStr, err := json.Marshal(o.Ages)
	panicErr(err)

	aliasStr, err := json.Marshal(o.Alias)
	panicErr(err)

	pricesStr, err := json.Marshal(o.Prices)
	panicErr(err)

	addressStr, err := json.Marshal(o.Address)
	panicErr(err)

	addressPtrStr, err := json.Marshal(o.AddressPtr)
	panicErr(err)

	addressesStr, err := json.Marshal(o.Addresses)
	panicErr(err)

	addressesPtrStr, err := json.Marshal(o.AddressesPtr)
	panicErr(err)

	return []any{o.Name, o.IsHuman, string(permsStr), string(agesStr), string(aliasStr), string(pricesStr), string(addressStr), string(addressPtrStr), string(addressesStr), string(addressesPtrStr), o.Id}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
func (o *Role) Scan(row *sql.Row) error {
	isHuman := 0
	var permsStr, agesStr, aliasStr, pricesStr, addressStr, addressPtrStr, addressesStr, addressesPtrStr []byte
	err := row.Scan(&o.Name, &isHuman, &permsStr, &agesStr, &aliasStr, &pricesStr, &addressStr, &addressPtrStr, &addressesStr, &addressesPtrStr, &o.Id)
	if err != nil {
		return err
	}

	o.IsHuman = isHuman > 0

	err = json.Unmarshal(permsStr, &o.Permissions)
	panicErr(err)

	err = json.Unmarshal(agesStr, &o.Ages)
	panicErr(err)

	err = json.Unmarshal(aliasStr, &o.Alias)
	panicErr(err)

	err = json.Unmarshal(pricesStr, &o.Prices)
	panicErr(err)

	err = json.Unmarshal(addressStr, &o.Address)
	panicErr(err)

	err = json.Unmarshal(addressPtrStr, &o.AddressPtr)
	panicErr(err)

	err = json.Unmarshal(addressesStr, &o.Addresses)
	panicErr(err)

	err = json.Unmarshal(addressesPtrStr, &o.AddressesPtr)
	panicErr(err)

	return nil
}

type Address struct {
	Street string
	City   string
	Zip    []string
}
