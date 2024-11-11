package request

type RegisterVerificationRequest struct {
	Type         string `form:"type" binding:"omitempty,oneof=forgotPwd unusualLogin"`
	Email        string `form:"email" binding:"required_without=MobileNumber,omitempty,email"`
	CountryCode  string `form:"countryCode" binding:"required_without=Email,required_with=MobileNumber,omitempty,number"`
	MobileNumber string `form:"mobileNumber" binding:"required_without=Email,required_with=CountryCode,omitempty,number"`
}

type VerificationRequest struct {
	Type                   string `json:"type" binding:"required,oneof=forgotPwd unusualLogin"`
	Email                  string `json:"email" binding:"required_without=MobileNumber,omitempty,email"`
	CountryCode            string `json:"countryCode" binding:"required_without=Email,required_with=MobileNumber,omitempty,number"`
	MobileNumber           string `json:"mobileNumber" binding:"required_without=Email,required_with=CountryCode,omitempty,number"`
	VerificationCodePrefix string `json:"verificationCodePrefix" binding:"required,alpha,len=3,uppercase"`
	VerificationCode       string `json:"verificationCode" binding:"required,number,len=6"`
	VerificationCodeToken  string `json:"verificationCodeToken" binding:"required"`
}
