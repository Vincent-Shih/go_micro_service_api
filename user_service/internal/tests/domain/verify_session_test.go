package domain_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"go_micro_service_api/user_service/internal/domain/vo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateVerificationCode(t *testing.T) {

	t.Run("check code", func(t *testing.T) {
		session := &vo.VerificationSession{}
		session.NextCode()

		assert.Equal(t, 3, len(session.Prefix))
		assert.Equal(t, 6, len(session.Code))
	})

	t.Run("check token", func(t *testing.T) {
		session := &vo.VerificationSession{}
		session.NextToken(context.Background(), "123", "456")

		hasher := md5.New()
		hasher.Write([]byte("123456"))
		expected := fmt.Sprintf("%x", hasher.Sum(nil))

		assert.Equal(t, expected, session.Token)
	})

	t.Run("check full code", func(t *testing.T) {
		session := vo.NewVerificationSession("", "ABC", "123456", "")

		assert.Equal(t, "ABC-123456", session.GetVerificationCode())
	})

	t.Run("verify success", func(t *testing.T) {
		session := vo.NewVerificationSession("", "ABC", "123456", "123456789")

		assert.True(t, session.Verify("ABC-123456", "123456789"))
	})

	t.Run("verify with type success", func(t *testing.T) {
		session := vo.NewVerificationSession("forgotPwd", "ABC", "123456", "123456789")

		assert.True(t, session.Verify("ABC-123456", "123456789"))
	})

	t.Run("verify fail", func(t *testing.T) {
		session := vo.NewVerificationSession("", "ABC", "123456", "123456789")

		assert.False(t, session.Verify("AB-3456", "123456789"))
		assert.False(t, session.Verify("ABC-123456", "126789"))
		assert.False(t, session.Verify("AB1456", "126789"))
	})
}
