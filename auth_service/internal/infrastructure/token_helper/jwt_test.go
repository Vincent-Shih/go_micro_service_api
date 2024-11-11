package token_helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	jwtToken := NewJwtToken()
	ctx := context.Background()
	secret := "secret"

	t.Run("Create token", func(t *testing.T) {
		claims := map[string]any{
			"email": "test1234@gamil.com",
			"role":  "admin",
			"name":  "test",
			"iat":   123456,
		}

		token, err := jwtToken.Create(ctx, secret, claims)
		assert.Nil(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestValidateToken(t *testing.T) {
	jwtToken := NewJwtToken()
	ctx := context.Background()
	secret := "secret"

	claims := map[string]any{
		"email": "test1234@gamil.com",
		"role":  "admin",
		"name":  "test",
		"iat":   123456,
	}

	token, err := jwtToken.Create(ctx, secret, claims)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	t.Run("Validate token", func(t *testing.T) {
		claims, err := jwtToken.Validate(ctx, token, secret)
		assert.Nil(t, err)
		assert.NotEmpty(t, claims)
	})

	t.Run("Validate token with invalid secret", func(t *testing.T) {
		claims, err := jwtToken.Validate(ctx, token, "invalid_secret")
		assert.NotNil(t, err)
		assert.Empty(t, claims)
	})
}

func TestGetPayload(t *testing.T) {
	JwtToken := NewJwtToken()
	ctx := context.Background()
	secret := "secret"

	claims := map[string]any{
		"email": "test1234@gamil.com",
		"role":  "admin",
		"name":  "test",
		"iat":   123456,
	}

	token, err := JwtToken.Create(ctx, secret, claims)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	t.Run("Get valid token", func(t *testing.T) {
		payload, err := JwtToken.GetPayload(ctx, token)
		assert.Nil(t, err)
		assert.NotEmpty(t, payload)
		assert.Equal(t, claims["email"], payload["email"])
		assert.Equal(t, claims["role"], payload["role"])
		assert.Equal(t, claims["name"], payload["name"])

		// Jwt token will convert number to float64
		// So we need to convert it back to int
		assert.Equal(t, claims["iat"].(int), int(payload["iat"].(float64)))
	})

	t.Run("Get invalid token", func(t *testing.T) {
		payload, err := JwtToken.GetPayload(ctx, "invalid_token")
		assert.NotNil(t, err)
		assert.Empty(t, payload)
	})
}
