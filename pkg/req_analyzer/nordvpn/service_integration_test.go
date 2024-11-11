package nordvpn_test

import (
	"context"
	"go_micro_service_api/pkg/req_analyzer/nordvpn"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-resty/resty/v2"
)

func TestGetLocation(t *testing.T) {
	type expectedType struct {
		IP          string
		Location    string
		Country     string
		CountryCode string
	}

	tests := []struct {
		name     string
		ip       string
		expected expectedType
	}{
		{
			name: "valid ip",
			ip:   "140.112.41.24",
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
			expected: expectedType{IP: ""},
		},
		{
			name:     "invalid ip",
			ip:       "ejwdsdsadadad==",
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

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
