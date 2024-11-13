package token_helper

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"

	"github.com/golang-jwt/jwt/v5"
)

type JwtToken struct{}

func NewJwtToken() *JwtToken {
	return &JwtToken{}
}

var _ TokenHelper = (*JwtToken)(nil)

func (j *JwtToken) Create(ctx context.Context, secret string, claims map[string]interface{}) (string, *cus_err.CusError) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		cus_err := cus_err.New(cus_err.InternalServerError, "Failed to create token", err)
		cus_otel.Error(ctx, cus_err.Message())
		return "", cus_err
	}
	return tokenString, nil
}

func (j *JwtToken) Validate(ctx context.Context, tokenString string, secret string) (map[string]interface{}, *cus_err.CusError) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			cusErr := cus_err.New(cus_err.InternalServerError, fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
			cus_otel.Error(ctx, cusErr.Message())
			return nil, cusErr
		}

		return []byte(secret), nil
	})

	if err != nil {
		cusErr := cus_err.New(cus_err.Unauthorized, "Jwt token is invalid", err)
		cus_otel.Error(ctx, cusErr.Message())
		return nil, cusErr
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	cusErr := cus_err.New(cus_err.Unauthorized, "Invalid token claims")
	cus_otel.Error(ctx, cusErr.Message())

	return nil, cusErr
}

func (j *JwtToken) GetPayload(ctx context.Context, tokenString string) (map[string]interface{}, *cus_err.CusError) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		cusErr := cus_err.New(cus_err.Unauthorized, "Jwt token is invalid", err)
		cus_otel.Error(ctx, cusErr.Message())
		return nil, cusErr
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	cusErr := cus_err.New(cus_err.Unauthorized, "Invalid token claims")
	cus_otel.Error(ctx, cusErr.Message())

	return nil, cusErr
}
