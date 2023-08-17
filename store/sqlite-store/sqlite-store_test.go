package sqlitestore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/yinloo-ola/srbac/store"
)

type Role struct {
	Name         string     `db:"name"`
	Permissions  []int64    `db:"permissions,json"`
	Ages         []int16    `db:"ages,json"`
	Alias        []string   `db:"alias,json"`
	Prices       []float32  `db:"prices,json"`
	Address      Address    `db:"address,json"`
	AddressPtr   *Address   `db:"addressPtr,json"`
	Addresses    []Address  `db:"addresses,json"`
	AddressesPtr []*Address `db:"addressesPtr,json"`
	Id           int64      `db:"id,pk"`
}

type Address struct {
	Street string
	City   string
	Zip    []string
}

func TestNew(t *testing.T) {
	path := "./rbac.db"
	roleStore, err := NewStore[Role](path)
	if err != nil {
		t.Fatalf("fail to create roleStore %v", err)
	}

	t.Cleanup(func() {
		errRemove := os.Remove(path)
		if errRemove != nil {
			t.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = roleStore.db.PingContext(ctx)
	if err != nil {
		t.Fatalf("ping fail %v", err)
	}
	role := Role{
		Name:         "admin",
		Permissions:  []int64{1, 2, 3},
		Alias:        []string{"a", "b"},
		Ages:         []int16{34, 22},
		Prices:       []float32{4.5, 3.2},
		Address:      Address{"street", "city", []string{"1", "2", "3"}},
		AddressPtr:   &Address{"streetPtr", "cityPtr", []string{"4", "5", "6"}},
		Addresses:    []Address{{"street1", "city1", []string{"7", "8", "9"}}, {"street2", "city2", []string{"10", "11", "12"}}},
		AddressesPtr: []*Address{{"streetPtr1", "cityPtr1", []string{"13", "14", "15"}}, {"streetPtr2", "cityPtr2", []string{"16", "17", "18"}}},
	}
	id, err := roleStore.Insert(role)
	if err != nil {
		t.Fatalf("fail to insert: %v", err)
	}
	role.Id = id
	if id != 1 {
		t.Errorf("expect role id to be 1 but gotten %v", id)
	}

	fmt.Println("id", id)

	role.Name = "super_admin"
	role.Permissions = []int64{4, 5, 6}
	err = roleStore.Update(id, role)
	if err != nil {
		t.Fatalf("fail to update %v", err)
	}

	err = roleStore.Update(100, role)
	if err != store.ErrNotFound {
		t.Fatalf("fail to update %v", err)
	}

	roleOut, err := roleStore.GetOne(id)
	if err != nil {
		t.Fatalf("GetOne failed: %v", err)
	}

	if !reflect.DeepEqual(role, roleOut) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	roleOut2, err := roleStore.GetOne(100)
	if err != store.ErrNotFound {
		t.Fatalf("expected error but gotten: %#v", roleOut2)
	}

}
