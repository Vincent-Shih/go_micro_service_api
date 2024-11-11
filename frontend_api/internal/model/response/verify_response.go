package response

// register verification session
type RegisterVerificationResponse struct {
	VerificationCodePrefix string `json:"verificationCodePrefix"`
	VerificationCodeToken  string `json:"verificationCodeToken"`
}

type VerificationResponse struct {
	Account      string `json:"account,omitempty"`
	Email        string `json:"email,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	MobileNumber string `json:"mobileNumber,omitempty"`
}

// if verification failed
type VerificationErrorResponse struct {
	// 已錯誤次數
	ErrorCount int `json:"errorCount"`
	// 總共幾次機會
	TotalAttempts int `json:"totalAttempts"`
}
