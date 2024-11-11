package vo

type UserList struct {
	UserId       int64  `json:"user_id"`
	Account      string `json:"account"`
	Email        string `json:"email"`
	CountryCode  string `json:"country_code"`
	MobileNumber string `json:"mobile_number"`
}
