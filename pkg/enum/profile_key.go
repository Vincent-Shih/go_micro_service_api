package enum

import "go_micro_service_api/pkg/cus_err"

type Profile struct {
	ID     int
	String string
}

var ProfileKey = struct {
	Email        Profile
	CountryCode  Profile
	MobileNumber Profile
	Account      Profile
}{
	Email: Profile{
		ID:     1,
		String: "Email",
	},
	CountryCode: Profile{
		ID:     2,
		String: "CountryCode",
	},
	MobileNumber: Profile{
		ID:     3,
		String: "MobileNumber",
	},
	Account: Profile{
		ID:     4,
		String: "Account",
	},
}

func ProfileKeyFromId(id int) (Profile, *cus_err.CusError) {
	switch id {
	case 1:
		return ProfileKey.Email, nil
	case 2:
		return ProfileKey.CountryCode, nil
	case 3:
		return ProfileKey.MobileNumber, nil
	case 4:
		return ProfileKey.Account, nil
	}

	return Profile{}, cus_err.New(cus_err.InvalidArgument, "invalid profile key")
}
