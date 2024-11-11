// Your good friend for development
// https://developers.google.com/oauthplayground
package google_test

import (
	"context"
	"go_micro_service_api/user_service/internal/infrastructure/client/google"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-resty/resty/v2"
)

func TestGetMe(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid token",
			token:    "", // find it from your good friend on top of this file
			expected: true,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
		{
			name:     "invalid token",
			token:    "ejwdsdsadadad==",
			expected: false,
		},
	}

	if tests[0].token == "" {
		t.Skip("please provide token")
	}

	client := resty.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			service := google.NewService(client)
			res, _ := service.GetMe(ctx, test.token)

			assert.Equal(t, res.ID != "", test.expected)
		})
	}
}
