package nordvpn

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"net"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const URL = "https://nordvpn.com"
const GetLocationPath = "/wp-admin/admin-ajax.php"

type Service struct {
	url    *url.URL
	client *resty.Client
}

func NewService(client *resty.Client) *Service {
	serviceUrl, _ := url.Parse(URL)
	return &Service{
		url:    serviceUrl,
		client: client,
	}
}

type GetLocationResponse struct {
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
	IP   string `json:"ip"`
	ISP  string `json:"isp"`
	Host struct {
		Domain    string `json:"domain"`
		IPAddress string `json:"ip_address"`
		PrefixLen int64  `json:"prefix_len"`
	} `json:"host"`
	Status      bool   `json:"status"`
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	Location    string `json:"location"`
	AreaCode    string `json:"area_code"`
	CountryCode string `json:"country_code"`
}

// GetLocation returns ip location
// Parameters:
//   - ip: ip address
//
// Returns:
//   - *GetLocationResponse: ip location
//   - *cus_err.CusError: error
func (s *Service) GetLocation(ctx context.Context, ip string) (*GetLocationResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	result := &GetLocationResponse{IP: ip}

	// 內部地址不查詢
	if s.isLocalhost(ip) || s.isInternalIP(ip) {
		return result, nil
	}

	// invalid ip
	if ip == "" || net.ParseIP(ip) == nil {
		return result, nil
	}

	url := s.url.JoinPath(GetLocationPath).String()
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"action": "get_user_info_data",
			"ip":     ip,
		}).
		SetResult(result).
		EnableTrace().
		Get(url)

	cus_otel.TraceRestyResponse(ctx, "nordvpn client trace info", url, resp)

	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError,
			fmt.Sprintf("Search ip: %s ,Occur error", ip), err)
		cus_otel.Error(ctx, cusErr.Error())
		return result, cusErr
	}

	return result, nil
}

func (s *Service) isLocalhost(ip string) bool {
	addrs, err := net.LookupIP(ip)
	if err != nil {
		return false
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil && (ipv4.IsLoopback() || ipv4.Equal(net.IPv4zero)) {
			return true
		}
	}

	return false
}

func (s *Service) isInternalIP(ip string) bool {
	if ipv4 := net.ParseIP(ip).To4(); ipv4 != nil {
		return ipv4.IsPrivate()
	}

	return false
}
