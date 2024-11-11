package response

type LoignPassResponse struct {
	Account      string `json:"account,omitempty"`
	Email        string `json:"email,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
	MobileNumber string `json:"mobileNumber,omitempty"`
}

// 判斷：會員使用從未登入過的瀏覽器，且IP地址變更到不同縣市，兩者缺一不可
type LoginAnomalousResponse struct {
	Email        string `json:"email"`
	CountryCode  string `json:"countryCode"`
	MobileNumber string `json:"mobileNumber"`
}

// if failed
type LoginErrorResponse struct {
	// 已錯誤次數
	ErrorCount int `json:"errorCount"`
	// 總共幾次機會
	TotalAttempts int `json:"totalAttempts"`
}
