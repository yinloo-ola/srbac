package sqlitestore

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	out := make([]any, 0, 11)
	out = append(out, o.Name, o.IsHuman)

	buffer := bytes.NewBuffer(make([]byte, 0, 500))
	enc := json.NewEncoder(buffer)

	err := enc.Encode(o.Permissions)
	panicErr(err)

	err = enc.Encode(o.Ages)
	panicErr(err)

	err = enc.Encode(o.Alias)
	panicErr(err)

	err = enc.Encode(o.Prices)
	panicErr(err)

	err = enc.Encode(o.Address)
	panicErr(err)

	err = enc.Encode(o.AddressPtr)
	panicErr(err)

	err = enc.Encode(o.Addresses)
	panicErr(err)

	err = enc.Encode(o.AddressesPtr)
	panicErr(err)

	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic("fail to read buffer: " + err.Error())
		}
		out = append(out, line)
	}

	out = append(out, o.Id)
	return out
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
func (o *Role) ScanRow(row RowScanner) error {
	isHuman := 0
	var permsStr, agesStr, aliasStr, pricesStr, addressStr, addressPtrStr, addressesStr, addressesPtrStr []byte
	err := row.Scan(&o.Name, &isHuman, &permsStr, &agesStr, &aliasStr, &pricesStr, &addressStr, &addressPtrStr, &addressesStr, &addressesPtrStr, &o.Id)
	if err != nil {
		return err
	}

	o.IsHuman = isHuman > 0

	total := len(permsStr) + len(agesStr) + len(aliasStr) + len(pricesStr) + len(addressStr) + len(addressPtrStr) + len(addressesStr) + len(addressesPtrStr)
	buffer := bytes.NewBuffer(make([]byte, 0, total))
	buffer.Write(permsStr)
	buffer.Write(agesStr)
	buffer.Write(aliasStr)
	buffer.Write(pricesStr)
	buffer.Write(addressStr)
	buffer.Write(addressPtrStr)
	buffer.Write(addressesStr)
	buffer.Write(addressesPtrStr)

	decoder := json.NewDecoder(buffer)

	err = decoder.Decode(&o.Permissions)
	panicErr(err)

	err = decoder.Decode(&o.Ages)
	panicErr(err)

	err = decoder.Decode(&o.Alias)
	panicErr(err)

	err = decoder.Decode(&o.Prices)
	panicErr(err)

	err = decoder.Decode(&o.Address)
	panicErr(err)

	err = decoder.Decode(&o.AddressPtr)
	panicErr(err)

	err = decoder.Decode(&o.Addresses)
	panicErr(err)

	err = decoder.Decode(&o.AddressesPtr)
	panicErr(err)

	return nil
}

type Address struct {
	Street string
	City   string
	Zip    []string
}
