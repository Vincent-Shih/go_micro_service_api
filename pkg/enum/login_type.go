package enum

import "go_micro_service_api/pkg/cus_err"

type LoginType struct {
	Id     int
	String string
}

var LoginTypes = struct {
	Account      LoginType
	Email        LoginType
	MobileNumber LoginType
}{
	Account: LoginType{
		Id:     1,
		String: "account",
	},
	Email: LoginType{
		Id:     2,
		String: "email",
	},
	MobileNumber: LoginType{
		Id:     3,
		String: "mobileNumber",
	},
}

func LoginTypeFromString(v string) (LoginType, *cus_err.CusError) {
	switch v {
	case "account":
		return LoginTypes.Account, nil
	case "email":
		return LoginTypes.Email, nil
	case "mobileNumber":
		return LoginTypes.MobileNumber, nil
	}
	return LoginType{}, cus_err.New(cus_err.InvalidArgument, "invalid login type")
}
