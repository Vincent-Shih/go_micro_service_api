package request

type LoginRequest struct {
	LoginType    string `json:"loginType" binding:"required,oneof=account email mobileNumber"`
	Account      string `json:"account" binding:"required_without_all=Email MobileNumber,omitempty,min=3,max=50,one_alpha,one_num"`
	Email        string `json:"email" binding:"required_without_all=Account MobileNumber,omitempty,email"`
	CountryCode  string `json:"countryCode" binding:"required_without_all=Account Email,required_with=MobileNumber,omitempty,number"`
	MobileNumber string `json:"mobileNumber" binding:"required_without_all=Account Email,required_with=CountryCode,omitempty,number"`
	Password     string `json:"password" binding:"required,min=6,max=15,one_alpha"`
}
