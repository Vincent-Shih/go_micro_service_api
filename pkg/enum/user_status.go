package enum

import (
	"go_micro_service_api/pkg/cus_err"
)

type UserStatus int

var UserStatusType = struct {
	Active UserStatus
	Locked UserStatus
}{
	Active: 1,
	Locked: 2,
}

func (s UserStatus) Int() int {
	return int(s)
}

func UserStatusFromInt(val int) (UserStatus, *cus_err.CusError) {
	switch val {
	case int(UserStatusType.Active):
		return UserStatusType.Active, nil
	case int(UserStatusType.Locked):
		return UserStatusType.Locked, nil
	default:
		return 0, cus_err.New(cus_err.AccountPasswordError, "invalid user status")
	}
}
