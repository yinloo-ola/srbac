package srbac

import (
	"fmt"

	"github.com/yinloo-ola/srbac/store"
)

type Role struct {
	Name        string  `db:"name"`
	Permissions []int64 `db:"permissions,json"`
	Id          int64   `db:"id,pk"`
}
type User struct {
	Id    int64   `db:"id,pk"`
	Roles []int64 `db:"roles,json"`
}
type Permission struct {
	Name string `db:"name"`
	Id   int64  `db:"id,pk"`
}
type Rbac struct {
	PermissionStore store.Store[Permission]
	RoleStore       store.Store[Role]
	UserStore       store.Store[User]
}

func New(permissionStore store.Store[Permission], roleStore store.Store[Role], userStore store.Store[User]) *Rbac {
	return &Rbac{
		PermissionStore: permissionStore,
		RoleStore:       roleStore,
		UserStore:       userStore,
	}
}

func (rbac *Rbac) HasPermission(userID int64, permissionID int64) (bool, error) {
	user, err := rbac.UserStore.GetOne(userID)
	if err != nil {
		return false, fmt.Errorf("rbac.UserStore.GetOne failed: %w", err)
	}

	roles, err := rbac.RoleStore.GetMulti(user.Roles)
	if err != nil {
		return false, fmt.Errorf("rbac.RoleStore.GetMulti failed: %w", err)
	}
	for _, r := range roles {
		for _, p := range r.Permissions {
			if p == permissionID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (rbac *Rbac) GetUserPermissions(userID int64) ([]Permission, error) {
	user, err := rbac.UserStore.GetOne(userID)
	if err != nil {
		return nil, fmt.Errorf("rbac.UserStore.GetOne failed: %w", err)
	}

	roles, err := rbac.RoleStore.GetMulti(user.Roles)
	if err != nil {
		return nil, fmt.Errorf("rbac.RoleStore.GetMulti failed: %w", err)
	}

	permissionIDs := make([]int64, 0, len(roles)*3)
	for _, r := range roles {
		permissionIDs = append(permissionIDs, r.Permissions...)
	}

	permissions, err := rbac.PermissionStore.GetMulti(permissionIDs)
	if err != nil {
		return nil, fmt.Errorf("rbac.PermissionStore.GetMulti failed: %w", err)
	}
	return permissions, nil
}
