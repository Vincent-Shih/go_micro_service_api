package nordvpn_test

import (
	"context"
	"go_micro_service_api/user_service/internal/infrastructure/client/nordvpn"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
)

func TestMockGetLocation(t *testing.T) {
	type expectedType struct {
		IP          string
		Location    string
		Country     string
		CountryCode string
	}

	tests := []struct {
		name     string
		ip       string
		respCode int
		resp     any
		expected expectedType
	}{
		{
			name:     "valid ip",
			ip:       "140.112.41.24",
			respCode: http.StatusOK,
			resp: map[string]any{
				"coordinates": map[string]any{
					"latitude":  25.0504,
					"longitude": 121.5324,
				},
				"ip":  "140.112.41.24",
				"isp": "Taiwan Academic Network",
				"host": map[string]any{
					"domain":     "ntu.edu.tw",
					"ip_address": "140.112.41.24",
					"prefix_len": 21,
				},
				"status":       false,
				"country":      "Taiwan",
				"region":       "Taipei City",
				"city":         "Taipei",
				"location":     "Taiwan, Taipei City, Taipei",
				"area_code":    "Unknown",
				"country_code": "TW",
			},
			expected: expectedType{
				IP:          "140.112.41.24",
				Location:    "Taiwan, Taipei City, Taipei",
				Country:     "Taiwan",
				CountryCode: "TW",
			},
		},
		{
			name:     "empty ip",
			ip:       "",
			respCode: http.StatusOK,
			resp: map[string]any{
				"coordinates": false,
				"ip":          "",
				"isp":         "Unknown",
				"host":        "Unknown",
				"status":      false,
				"location":    "Unknown",
			},
			expected: expectedType{IP: ""},
		},
		{
			name:     "invalid ip",
			ip:       "ejwdsdsadadad==",
			respCode: http.StatusOK,
			resp: map[string]any{
				"coordinates": false,
				"ip":          "ejwdsdsadadad==",
				"isp":         "Unknown",
				"host":        "Unknown",
				"status":      false,
				"location":    "Unknown",
			},
			expected: expectedType{IP: "ejwdsdsadadad=="},
		},
		{
			name:     "localhost",
			ip:       "localhost",
			expected: expectedType{IP: "localhost"},
		},
		{
			name:     "0.0.0.0",
			ip:       "0.0.0.0",
			expected: expectedType{IP: "0.0.0.0"},
		},
		{
			name:     "internal ip",
			ip:       "192.168.0.5",
			expected: expectedType{IP: "192.168.0.5"},
		},
	}

	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			httpmock.RegisterResponder(
				http.MethodGet,
				nordvpn.URL+nordvpn.GetLocationPath+"?action=get_user_info_data&ip="+test.ip,
				httpmock.NewJsonResponderOrPanic(test.respCode, test.resp),
			)

			service := nordvpn.NewService(client)
			res, _ := service.GetLocation(ctx, test.ip)

			assert.Equal(t, expectedType{
				IP:          res.IP,
				Location:    res.Location,
				Country:     res.Country,
				CountryCode: res.CountryCode,
			}, test.expected)
		})
	}
}
