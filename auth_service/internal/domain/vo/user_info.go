package vo

import "go_micro_service_api/pkg/enum"

type UserInfo struct {
	Id       int64
	Account  string
	Password string
	Status   enum.UserStatus
}
