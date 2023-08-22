package srbac

import (
	"fmt"

	"github.com/yinloo-ola/srbac/models"
	"github.com/yinloo-ola/srbac/store"
)

type Rbac struct {
	PermissionStore store.Store[models.Permission]
	RoleStore       store.Store[models.Role]
	UserStore       store.Store[models.User]
}

func NewRbac(permissionStore store.Store[models.Permission], roleStore store.Store[models.Role], userStore store.Store[models.User]) *Rbac {
	return &Rbac{
		PermissionStore: permissionStore,
		RoleStore:       roleStore,
		UserStore:       userStore,
	}
}

func (rbac *Rbac) HasPermission(userID string, permissionID int64) (bool, error) {
	users, err := rbac.UserStore.FindField("user_id", userID)
	if err != nil {
		return false, fmt.Errorf("rbac.UserStore.FindField failed: %w", err)
	}
	if len(users) != 1 {
		return false, store.ErrNotFound
	}

	roles, err := rbac.RoleStore.GetMulti(users[0].Roles)
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

func (rbac *Rbac) GetUserPermissions(userID string) ([]models.Permission, error) {
	users, err := rbac.UserStore.FindField("user_id", userID)
	if err != nil {
		return nil, fmt.Errorf("rbac.UserStore.FindField failed: %w", err)
	}
	if len(users) != 1 {
		return nil, store.ErrNotFound
	}

	roles, err := rbac.RoleStore.GetMulti(users[0].Roles)
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
