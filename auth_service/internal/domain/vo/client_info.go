package vo

import "go_micro_service_api/pkg/enum"

type ClientInfo struct {
	Id               int64
	MerchantId       int64
	ClientType       enum.Client
	LoginFailedTimes int
	TokenExpireSecs  int
	Active           bool
}
