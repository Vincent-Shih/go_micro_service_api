package helper_test

import (
	"go_micro_service_api/pkg/helper"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/go-playground/validator/v10"
)

func TestVerifiedCodeValidation(t *testing.T) {
	var validate = validator.New()
	_ = validate.RegisterValidation("verify_code", helper.VerifiedCodeValidation)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Valid code",
			value: "ABC-123456",
			want:  true,
		},
		{
			name:  "Invalid prefix case",
			value: "abc-123456",
			want:  false,
		},
		{
			name:  "Invalid prefix length",
			value: "ab-123456",
			want:  false,
		},
		{
			name:  "Invalid code length",
			value: "ABC-12345",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, validate.Var(tt.value, "verify_code") == nil)
		})
	}
}

func TestOneAlphaValidation(t *testing.T) {
	var validate = validator.New()
	_ = validate.RegisterValidation("one_alpha", helper.ContainsAtLeastOneAlpha)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Valid upper",
			value: "A132314",
			want:  true,
		},
		{
			name:  "Valid lower",
			value: "a132314",
			want:  true,
		},
		{
			name:  "Invalid",
			value: "123456",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, validate.Var(tt.value, "one_alpha") == nil)
		})
	}
}

func TestOneNumValidation(t *testing.T) {
	var validate = validator.New()
	_ = validate.RegisterValidation("one_num", helper.ContainsAtLeastOneNum)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Valid",
			value: "123456",
			want:  true,
		},
		{
			name:  "Invalid",
			value: "abc",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, validate.Var(tt.value, "one_num") == nil)
		})
	}
}
