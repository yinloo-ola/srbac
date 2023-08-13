package sqlitestore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

type Role struct {
	Name        string  `db:"name"`
	Permissions []int64 `db:"permissions,json"`
	Id          int64   `db:"id,pk"`
}

func TestNew(t *testing.T) {
	path := "./rbac.db"
	store, err := NewStore[Role](path)
	if err != nil {
		t.Fatalf("fail to create store %v", err)
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
	err = store.db.PingContext(ctx)
	if err != nil {
		t.Fatalf("ping fail %v", err)
	}
	role := Role{
		Name:        "admin",
		Permissions: []int64{1, 2, 3},
		Id:          5,
	}
	id, err := store.Insert(role)
	if err != nil {
		t.Fatalf("fail to insert: %v", err)
	}
	if id != 1 {
		t.Errorf("expect role id to be 1 but gotten %v", id)
	}

	fmt.Println("id", id)

	role.Name = "super_admin"
	role.Permissions = []int64{4, 5, 6}
	err = store.Update(id, role)
	if err != nil {
		t.Fatalf("fail to update %v", err)
	}
}
