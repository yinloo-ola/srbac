package sqlitestore

import (
	"context"
	"os"
	"testing"
	"time"
)

// func BenchmarkReflectGetOne(b *testing.B) {
// 	path := "./rbac.db"
// 	roleStore, err := NewStoreReflect[Role](path)
// 	if err != nil {
// 		b.Fatalf("fail to create roleStore %v", err)
// 	}

// 	b.Cleanup(func() {
// 		errRemove := os.Remove(path)
// 		if errRemove != nil {
// 			b.Fatalf("fail to clean up rbac.db. please clean up manually")
// 		}
// 	})

// 	ctx := context.Background()
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()
// 	err = roleStore.db.PingContext(ctx)
// 	if err != nil {
// 		b.Fatalf("ping fail %v", err)
// 	}
// 	role := Role{
// 		Name:         "admin",
// 		IsHuman:      true,
// 		Permissions:  []int64{1, 2, 3},
// 		Alias:        []string{"a", "b"},
// 		Ages:         []int16{34, 22},
// 		Prices:       []float32{4.5, 3.2},
// 		Address:      Address{"street", "city", []string{"1", "2", "3"}},
// 		AddressPtr:   &Address{"streetPtr", "cityPtr", []string{"4", "5", "6"}},
// 		Addresses:    []Address{{"street1", "city1", []string{"7", "8", "9"}}, {"street2", "city2", []string{"10", "11", "12"}}},
// 		AddressesPtr: []*Address{{"streetPtr1", "cityPtr1", []string{"13", "14", "15"}}, {"streetPtr2", "cityPtr2", []string{"16", "17", "18"}}},
// 	}
// 	id, err := roleStore.Insert(role)
// 	if err != nil {
// 		b.Fatalf("fail to insert: %v", err)
// 	}

// 	b.ResetTimer()
// 	var r Role
// 	for i := 0; i < b.N; i++ {
// 		r, err = roleStore.GetOne(id)
// 		if err != nil {
// 			b.Fatalf("fail to get %d: %s", id, err)
// 		}
// 	}
// 	_ = r
// }

func BenchmarkJsonGetOne(b *testing.B) {
	path := "rbac2.db"
	roleStore, err := NewStore[Role](path)
	if err != nil {
		b.Fatalf("fail to create roleStore %v", err)
	}

	b.Cleanup(func() {
		errRemove := os.Remove(path)
		if errRemove != nil {
			b.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
		errRemove = os.Remove(path + "-shm")
		if errRemove != nil {
			b.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
		errRemove = os.Remove(path + "-wal")
		if errRemove != nil {
			b.Fatalf("fail to clean up rbac.db. please clean up manually")
		}
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = roleStore.db.PingContext(ctx)
	if err != nil {
		b.Fatalf("ping fail %v", err)
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
	id, err := roleStore.Insert(role)
	if err != nil {
		b.Fatalf("fail to insert: %v", err)
	}

	b.ResetTimer()

	var r Role
	for i := 0; i < b.N; i++ {
		r, err = roleStore.GetOne(id)
		if err != nil {
			b.Fatalf("fail to get %d: %s", id, err)
		}
	}
	_ = r
}
