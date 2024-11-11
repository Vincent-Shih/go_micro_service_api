package entity

type Profile struct {
	Account      string
	Email        string
	CountryCode  string
	MobileNumber string
	ThirdParties ThridParties
}

type ThridParties struct {
	Google  ThirdParty
	Meta    ThirdParty
	Twitter ThirdParty
	LINE    ThirdParty
}

type ThirdParty struct {
	ID string
}
