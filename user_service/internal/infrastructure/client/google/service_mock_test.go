// Your good friend for development
// https://developers.google.com/oauthplayground
package google_test

import (
	"context"
	"go_micro_service_api/user_service/internal/infrastructure/client/google"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
)

func TestMockGetMe(t *testing.T) {
	tests := []struct {
		name     string
		respCode int
		resp     any
		expected bool
	}{
		{
			name:     "valid token",
			respCode: http.StatusOK,
			resp: map[string]any{
				"issued_to":      "111111111111.apps.googleusercontent.com",
				"user_id":        "111111111111111111111",
				"expires_in":     2753,
				"access_type":    "offline",
				"audience":       "111111111111.apps.googleusercontent.com",
				"scope":          "https://www.googleapis.com/auth/userinfo.email openid",
				"email":          "xxx@cus.go",
				"verified_email": true,
			},
			expected: true,
		},
		{
			name:     "empty token",
			respCode: http.StatusBadRequest,
			resp: map[string]any{
				"error":             "invalid_token",
				"error_description": "Either access_token, id_token, or token_handle required",
			},
			expected: false,
		},
		{
			name:     "invalid token",
			respCode: http.StatusBadRequest,
			resp: map[string]any{
				"error":             "invalid_token",
				"error_description": "Invalid Value",
			},
			expected: false,
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
				google.URL+google.PathGetMe,
				httpmock.NewJsonResponderOrPanic(test.respCode, test.resp),
			)

			service := google.NewService(client)
			res, _ := service.GetMe(ctx, "")

			assert.Equal(t, res.ID != "", test.expected)
		})
	}
}
