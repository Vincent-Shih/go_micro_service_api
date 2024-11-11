package enum

import "go_micro_service_api/pkg/cus_err"

type Permission struct {
	Id   int64
	Name string
}

var PermissionType = struct {
	Withdraw Permission
	Deposit  Permission
	PlayGame Permission
}{
	Withdraw: Permission{
		Id:   1,
		Name: "PlayerWithdraw",
	},
	Deposit: Permission{
		Id:   2,
		Name: "PlayerDeposit",
	},
	PlayGame: Permission{
		Id:   3,
		Name: "PlayerPlayGame",
	},
}

func PermissionById(id int64) (Permission, *cus_err.CusError) {
	switch id {
	case PermissionType.Withdraw.Id:
		return PermissionType.Withdraw, nil
	case PermissionType.Deposit.Id:
		return PermissionType.Deposit, nil
	case PermissionType.PlayGame.Id:
		return PermissionType.PlayGame, nil
	default:
		return Permission{},
			cus_err.New(cus_err.AccountPasswordError, "invalid permission id")
	}
}
