package srbac

import (
	"testing"

	"github.com/yinloo-ola/srbac/helper"
	"github.com/yinloo-ola/srbac/models"
	sqlitestore "github.com/yinloo-ola/srbac/store/sqlite-store"
)

func TestNew(t *testing.T) {
	path := "rbac.db"
	permissionStore, err := sqlitestore.NewStore[models.Permission](path)
	roleStore, err := sqlitestore.NewStore[models.Role](path)
	userStore, err := sqlitestore.NewStore[models.User](path)
	helper.PanicErr(err)
	rbac := NewRbac(
		permissionStore, roleStore, userStore,
	)
	_ = rbac
}
