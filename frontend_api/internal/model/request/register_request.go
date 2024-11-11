package request

type RegisterRequest struct {
	Account                string `json:"account" binding:"required,min=3,max=50,one_alpha,one_num"`
	Email                  string `json:"email" binding:"required_without=MobileNumber,omitempty,email"`
	CountryCode            string `json:"countryCode" binding:"required_without=Email,required_with=MobileNumber,omitempty,number"`
	MobileNumber           string `json:"mobileNumber" binding:"required_without=Email,required_with=CountryCode,omitempty,number"`
	Password               string `json:"password" binding:"required,min=6,max=15,one_alpha"`
	VerificationCodePrefix string `json:"verificationCodePrefix" binding:"required,alpha,len=3,uppercase"`
	VerificationCode       string `json:"verificationCode" binding:"required,number,len=6"`
	VerificationCodeToken  string `json:"verificationCodeToken" binding:"required"`
}

type CheckUserExistenceRequest struct {
	Account      string `form:"account" binding:"required_without_all=Email MobileNumber" example:"account"`
	Email        string `form:"email" binding:"required_without_all=Account MobileNumber,omitempty,email" example:"test@example.com"`
	CountryCode  string `form:"countryCode" binding:"required_with=MobileNumber,omitempty,number" example:"886"`
	MobileNumber string `form:"mobileNumber" binding:"required_without_all=Account Email,required_with=CountryCode,omitempty,number" example:"912345678"`
}
