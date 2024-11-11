package enum

import "go_micro_service_api/pkg/cus_err"

type Provider struct {
	ID     int
	String string
}

var OAuthProvider = struct {
	Google  Provider
	Meta    Provider
	Twitter Provider
	LINE    Provider
}{
	Google: Provider{
		ID:     1,
		String: "Google",
	},
	Meta: Provider{
		ID:     2,
		String: "Meta",
	},
	Twitter: Provider{
		ID:     3,
		String: "Twitter",
	},
	LINE: Provider{
		ID:     4,
		String: "LINE",
	},
}

func OAuthProviderFromId(id int) (Provider, *cus_err.CusError) {
	switch id {
	case 1:
		return OAuthProvider.Google, nil
	case 2:
		return OAuthProvider.Meta, nil
	case 3:
		return OAuthProvider.Twitter, nil
	case 4:
		return OAuthProvider.LINE, nil
	}

	return Provider{}, cus_err.New(cus_err.InvalidArgument, "invalid oauth provider")
}

func OAuthProviderFromString(value string) (Provider, *cus_err.CusError) {
	switch value {
	case "Google":
		return OAuthProvider.Google, nil
	case "Meta":
		return OAuthProvider.Meta, nil
	case "Twitter":
		return OAuthProvider.Twitter, nil
	case "LINE":
		return OAuthProvider.LINE, nil
	}

	return Provider{}, cus_err.New(cus_err.InvalidArgument, "invalid oauth provider")
}
