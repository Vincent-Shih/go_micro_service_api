package enum

import "go_micro_service_api/pkg/cus_err"

type Client struct {
	Id     int
	String string
}

var ClientType = struct {
	Frontend Client
	Backend  Client
}{
	Frontend: Client{
		Id:     1,
		String: "Frontend",
	},
	Backend: Client{
		Id:     2,
		String: "Backend",
	},
}

func ClientTypeFromId(id int) (Client, *cus_err.CusError) {
	switch id {
	case 1:
		return ClientType.Frontend, nil
	case 2:
		return ClientType.Backend, nil
	}
	return Client{}, cus_err.New(cus_err.AccountPasswordError, "invalid client type")
}
