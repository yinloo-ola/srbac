package sqlitestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/yinloo-ola/srbac/store"
)

func TestNew(t *testing.T) {
	path := "rbac.db"
	now := time.Now()
	roleStore, err := NewStore[Role](path)
	if err != nil {
		t.Fatalf("fail to create roleStore %v", err)
	}
	fmt.Printf("new store duration: %s\n", time.Since(now))

	t.Cleanup(func() {
		errRemove := os.Remove(path)
		if errRemove != nil {
			t.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
		errRemove = os.Remove(path + "-shm")
		if errRemove != nil {
			t.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
		errRemove = os.Remove(path + "-wal")
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
		IsHuman:      true,
		Permissions:  []int64{1, 2, 3},
		Alias:        []string{"a", "b"},
		Ages:         []int16{34, 22},
		Prices:       []float32{4.5, 3.2},
		Address:      Address{"street", "city", []string{"1", "2", "3"}},
		AddressPtr:   &Address{"streetPtr", "cityPtr", []string{"4", "5", "6"}},
		Addresses:    []Address{{"street1", "city1", []string{"7", "8", "9"}}, {"street2", "city2", []string{"10", "11", "12"}}},
		AddressesPtr: []*Address{{"streetPtr1", "cityPtr1", []string{"13", "14", "15"}}, {"streetPtr2", "cityPtr2", []string{"16", "17", "18"}}},
	}
	now = time.Now()
	id, err := roleStore.Insert(role)
	if err != nil {
		t.Fatalf("fail to insert: %v", err)
	}
	role.Id = id
	if id != 1 {
		t.Errorf("expect role id to be 1 but gotten %v", id)
	}
	fmt.Printf("insert duration: %s\n", time.Since(now))

	role.Name = "super_admin"
	role.Permissions = []int64{4, 5, 6}
	now = time.Now()
	err = roleStore.Update(id, role)
	if err != nil {
		t.Fatalf("fail to update %v", err)
	}
	fmt.Printf("update duration: %s\n", time.Since(now))

	now = time.Now()
	err = roleStore.Update(100, role)
	if err != store.ErrNotFound {
		t.Fatalf("fail to update %v", err)
	}
	fmt.Printf("update failed duration: %s\n", time.Since(now))

	now = time.Now()
	roleOut, err := roleStore.GetOne(id)
	if err != nil {
		t.Fatalf("GetOne failed: %v", err)
	}
	fmt.Println("duration [GetOne]", time.Since(now))

	if !reflect.DeepEqual(role, roleOut) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	roleOut2, err := roleStore.GetOne(100)
	if err != store.ErrNotFound {
		t.Fatalf("expected error but gotten: %#v", roleOut2)
	}
	fmt.Printf("GetOne not found duration: %s\n", time.Since(now))

	role2 := Role{
		Name:         "referee",
		IsHuman:      false,
		Permissions:  []int64{11, 22, 33},
		Alias:        []string{"aa", "bb"},
		Ages:         []int16{134, 222},
		Prices:       []float32{54.5, 53.2},
		Address:      Address{"2street", "2city", []string{"1", "2", "3", "22"}},
		AddressPtr:   &Address{"2streetPtr", "cityPtr", []string{"4", "5", "6"}},
		Addresses:    []Address{{"2street1", "city1", []string{"7", "8", "9", "22"}}, {"2street2", "city2", nil}},
		AddressesPtr: []*Address{{"streetPtr1", "2cityPtr1", []string{"13", "14", "15"}}, {"2streetPtr2", "cityPtr2", []string{"16", "17", "18"}}},
	}
	now = time.Now()
	id, err = roleStore.Insert(role2)
	if err != nil {
		t.Fatalf("fail to insert: %v", err)
	}
	fmt.Printf("insert duration: %s\n", time.Since(now))
	role2.Id = id
	if id != 2 {
		t.Errorf("expect role id to be 1 but gotten %v", id)
	}

	now = time.Now()
	rolesOut, err := roleStore.GetMulti([]int64{1, 2})
	if err != nil {
		t.Fatalf("GetMulti failed: %v", err)
	}
	fmt.Printf("getmulti duration: %s\n", time.Since(now))
	if len(rolesOut) != 2 {
		t.Fatalf("GetMulti returned wrong number of roles")
	}
	if !reflect.DeepEqual(role, rolesOut[0]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}
	if !reflect.DeepEqual(role2, rolesOut[1]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	rolesOut2, err := roleStore.GetMulti([]int64{3, 1})
	if err != nil {
		t.Fatalf("GetMulti failed: %v", err)
	}
	fmt.Printf("getmulti duration: %s\n", time.Since(now))
	if len(rolesOut2) != 1 {
		t.Fatalf("GetMulti returned wrong number of roles")
	}
	if !reflect.DeepEqual(role, rolesOut2[0]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	rolesOutAll, err := roleStore.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	fmt.Printf("GetAll duration: %s\n", time.Since(now))
	if len(rolesOutAll) != 2 {
		t.Fatalf("GetAll returned wrong number of roles")
	}
	if !reflect.DeepEqual(role, rolesOutAll[0]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}
	if !reflect.DeepEqual(role2, rolesOutAll[1]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	err = roleStore.DeleteMulti([]int64{3, 4})
	if err == nil {
		t.Fatalf("expected not found")
	}
	fmt.Printf("DeleteMulti duration: %s\n", time.Since(now))
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected not found but gotten %s", err)
	}

	now = time.Now()
	err = roleStore.DeleteMulti([]int64{2, 4})
	if err != nil {
		t.Fatalf("delete multi failed")
	}
	fmt.Printf("DeleteMulti duration: %s\n", time.Since(now))

	rolesOutAll2, err := roleStore.GetAll()
	if err != nil {
		t.Fatalf("roleStore.GetAll failed %s", err)
	}
	if len(rolesOutAll2) != 1 {
		t.Fatalf("roleStore.GetAll expect len 1, but got %d", len(rolesOutAll2))
	}
	if !reflect.DeepEqual(role, rolesOutAll[0]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	rolesFind, err := roleStore.FindField("name", "super_admin")
	if err != nil {
		t.Fatalf("roleStore.FindField failed %s", err)
	}
	fmt.Printf("FindField duration: %s\n", time.Since(now))
	if len(rolesFind) != 1 {
		t.Fatalf("roleStore.FindField expect len 1, but got %d", len(rolesFind))
	}
	if !reflect.DeepEqual(role, rolesFind[0]) {
		t.Errorf("expected role:%#v. gotten:%#v", role, roleOut)
	}

	now = time.Now()
	rolesFind2, err := roleStore.FindField("name", "admin")
	if err != nil {
		t.Fatalf("roleStore.FindField failed %s", err)
	}
	if len(rolesFind2) != 0 {
		t.Fatalf("roleStore.FindField should return nothing %s", err)
	}

	fmt.Printf("FindField duration: %s\n", time.Since(now))

	_, err = roleStore.Insert(Role{
		Name: "super_admin",
	})
	if err == nil {
		t.Fatalf("should get duplicate key error")
	}

}
