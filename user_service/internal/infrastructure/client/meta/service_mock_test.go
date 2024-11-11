// Your good friend for development
// https://developers.facebook.com/tools/explorer?version=v19.0
package meta_test

import (
	"context"
	"go_micro_service_api/user_service/internal/infrastructure/client/meta"
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
				"id":    "32213231131230187459",
				"email": "xxx@kgs.go",
			},
			expected: true,
		},
		{
			name:     "empty token",
			respCode: http.StatusBadRequest,
			resp: map[string]map[string]any{
				"error": {
					"message":    "An active access token must be used to query information about the current user.",
					"type":       "OAuthException",
					"code":       2500,
					"fbtrace_id": "AjYc7BDNFe8TcestAz9MQ_W",
				},
			},
			expected: false,
		},
		{
			name:     "invalid token",
			respCode: http.StatusBadRequest,
			resp: map[string]map[string]any{
				"error": {
					"message":    "The access token could not be decrypted",
					"type":       "OAuthException",
					"code":       190,
					"fbtrace_id": "AvwR8KibTLFSS6V50TyWMUl",
				},
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
				meta.URL+meta.PathGetMe,
				httpmock.NewJsonResponderOrPanic(test.respCode, test.resp),
			)

			service := meta.NewService(client)
			res, _ := service.GetMe(ctx, "")

			assert.Equal(t, res.ID != "", test.expected)
		})
	}
}
