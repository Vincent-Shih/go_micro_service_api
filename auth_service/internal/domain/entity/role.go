package entity

import (
	"go_micro_service_api/pkg/enum"
)

type Role struct {
	Id          int64             // Id is the unique identifier of the role.
	Name        string            // Name is the name of the role.
	Permissions []enum.Permission // Permissions is the list of Permission enums associated with the role.
	ClientType  enum.Client       // clientType is the type of client that the role is associated with.
	isSystem    bool              // isSystem is a flag indicating whether the role is created by default. merchant users can not modify system roles.
}

// IsSystem checks if the role is a system role by comparing its ID with the IDs of SystemRoles.
//
// Returns:
//   - true if the role's ID matches any of the SystemRoles' IDs.
//   - false otherwise.
func (r *Role) IsSystem() bool {
	allRoles := append(AllBackendRoles, AllFrontendRoles...)
	for _, role := range allRoles {
		if r.Id == role.Id {
			return true
		}
	}
	return false
}

func (r *Role) GetPermissionIds() []int64 {
	permissionIds := make([]int64, 0)
	for _, permission := range r.Permissions {
		permissionIds = append(permissionIds, permission.Id)
	}
	return permissionIds
}

var AllFrontendRoles = []Role{
	FrontendRoles.Guest,
	FrontendRoles.Player,
	FrontendRoles.KycVerifiedPlayer,
}

var AllBackendRoles = []Role{
	BackendRoles.Admin,
	BackendRoles.CustomerSupport,
}

// FrontendRoles contains the default frontend roles.
// These roles are create by default can't be modified by merchant users.
// FrontendRole id is start from 1
var FrontendRoles = struct {
	Guest             Role
	Player            Role
	KycVerifiedPlayer Role
}{
	Guest: Role{
		Id:          1,
		Name:        "Guest",
		Permissions: []enum.Permission{},
		isSystem:    true,
		ClientType:  enum.ClientType.Frontend,
	},
	Player: Role{
		Id:          2,
		Name:        "Player",
		Permissions: []enum.Permission{enum.PermissionType.PlayGame},
		isSystem:    true,
		ClientType:  enum.ClientType.Frontend,
	},
	KycVerifiedPlayer: Role{
		Id:   3,
		Name: "KycVerifiedPlayer",
		Permissions: []enum.Permission{
			enum.PermissionType.Deposit,
			enum.PermissionType.Withdraw,
			enum.PermissionType.PlayGame,
		},
		isSystem:   true,
		ClientType: enum.ClientType.Frontend,
	},
}

// BackendRoles contains the default backend roles.
// These roles are create by default can't be modified by merchant users.
// BackendRole id  is start from 101
var BackendRoles = struct {
	Admin           Role
	CustomerSupport Role
}{
	Admin: Role{
		Id:          101,
		Name:        "Admin",
		Permissions: []enum.Permission{}, // TODO: Add more backend permissions here
		isSystem:    true,
		ClientType:  enum.ClientType.Backend,
	},
	CustomerSupport: Role{
		Id:          102,
		Name:        "CustomerSupport",
		Permissions: []enum.Permission{}, // TODO: Add more backend permissions here
		isSystem:    true,
		ClientType:  enum.ClientType.Backend,
	},
}
