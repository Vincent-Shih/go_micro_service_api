package response

// if register success
type RegisterResponse struct {
	Account      string `json:"account"`
	Email        string `json:"email"`
	CountryCode  string `json:"countryCode"`
	MobileNumber string `json:"mobileNumber"`
}

type ExistenceResponse struct {
	Exists bool `json:"isExist" example:"false"`
}
