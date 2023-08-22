package srbac

import (
	"fmt"
	"testing"

	"github.com/yinloo-ola/srbac/helper"
	"github.com/yinloo-ola/srbac/models"
	sqlitestore "github.com/yinloo-ola/srbac/store/sqlite-store"
)

func TestNew(t *testing.T) {
	path := "rbac.db"
	permissionStore, err := sqlitestore.NewStore[models.Permission](path)
	helper.PanicErr(err)
	roleStore, err := sqlitestore.NewStore[models.Role](path)
	helper.PanicErr(err)
	userStore, err := sqlitestore.NewStore[models.User](path)
	helper.PanicErr(err)
	rbac := NewRbac(
		permissionStore, roleStore, userStore,
	)
	id, err := rbac.PermissionStore.Insert(models.Permission{
		Name:        "permissions.read",
		Description: "access to read permissions",
	})
	helper.PanicErr(err)
	fmt.Println(id)
	err = rbac.Close()
	helper.PanicErr(err)
}
