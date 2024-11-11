package aggregate

import "go_micro_service_api/user_service/internal/domain/entity"

type User struct {
	ID      int64
	Profile entity.Profile
}
