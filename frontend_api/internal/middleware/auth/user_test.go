package auth

import (
	"go_micro_service_api/pkg/enum"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewUserInfo(t *testing.T) {
	permissions := []enum.Permission{enum.PermissionType.Deposit, enum.PermissionType.Withdraw}
	clientId := int64(1)
	merchantId := int64(2)
	userAccount := "test"
	userId := int64(3)

	userInfo := NewUserInfo(permissions, clientId, merchantId, &userAccount, &userId)

	assert.Equal(t, permissions, userInfo.permissions)
	assert.Equal(t, clientId, userInfo.clientId)
	assert.Equal(t, merchantId, userInfo.merchantId)
	assert.Equal(t, &userAccount, userInfo.userAccount)
	assert.Equal(t, &userId, userInfo.userId)
}

func TestGetUserInfo(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)

	t.Run("UserInfoNotSet", func(t *testing.T) {
		userInfo, ok := GetUserInfo(c)
		assert.False(t, ok)
		assert.Nil(t, userInfo)
	})

	t.Run("UserInfoSetCorrectly", func(t *testing.T) {
		expectedUserInfo := &UserInfo{}
		c.Set(userInfoKey, expectedUserInfo)

		userInfo, ok := GetUserInfo(c)
		assert.True(t, ok)
		assert.Equal(t, expectedUserInfo, userInfo)
	})

	t.Run("UserInfoIncorrectType", func(t *testing.T) {
		c.Set(userInfoKey, "Not a UserInfo")

		userInfo, ok := GetUserInfo(c)
		assert.False(t, ok)
		assert.Nil(t, userInfo)
	})
}

func TestHasPermission(t *testing.T) {
	userInfo := &UserInfo{
		permissions: []enum.Permission{enum.PermissionType.Deposit, enum.PermissionType.Withdraw},
	}

	assert.True(t, userInfo.HasPermission(enum.PermissionType.Deposit))
	assert.True(t, userInfo.HasPermission(enum.PermissionType.Withdraw))
	assert.True(t, userInfo.HasPermission(enum.PermissionType.Deposit, enum.PermissionType.Withdraw))
	assert.False(t, userInfo.HasPermission(enum.PermissionType.Deposit, enum.PermissionType.Withdraw, enum.PermissionType.PlayGame))
	assert.False(t, userInfo.HasPermission(enum.PermissionType.PlayGame))
}

func TestGetPermissions(t *testing.T) {
	permissions := []enum.Permission{enum.PermissionType.Deposit, enum.PermissionType.Withdraw}
	userInfo := &UserInfo{permissions: permissions}

	assert.Equal(t, permissions, userInfo.GetPermissions())

	// Ensure that modifying the returned slice doesn't affect the original
	returnedPerms := userInfo.GetPermissions()
	returnedPerms[0] = enum.PermissionType.PlayGame
	assert.NotEqual(t, returnedPerms, userInfo.permissions)
}

func TestGetUserId(t *testing.T) {
	t.Run("UserIdSet", func(t *testing.T) {
		userId := int64(123)
		userInfo := &UserInfo{userId: &userId}

		id, ok := userInfo.GetUserId()
		assert.True(t, ok)
		assert.Equal(t, userId, id)
	})

	t.Run("UserIdNotSet", func(t *testing.T) {
		userInfo := &UserInfo{}

		id, ok := userInfo.GetUserId()
		assert.False(t, ok)
		assert.Equal(t, int64(0), id)
	})
}

func TestGetUserAccount(t *testing.T) {
	t.Run("UserAccountSet", func(t *testing.T) {
		account := "test@example.com"
		userInfo := &UserInfo{userAccount: &account}

		acc, ok := userInfo.GetUserAccount()
		assert.True(t, ok)
		assert.Equal(t, account, acc)
	})

	t.Run("UserAccountNotSet", func(t *testing.T) {
		userInfo := &UserInfo{}

		acc, ok := userInfo.GetUserAccount()
		assert.False(t, ok)
		assert.Equal(t, "", acc)
	})
}

func TestGetClientId(t *testing.T) {
	clientId := int64(456)
	userInfo := &UserInfo{clientId: clientId}
	assert.Equal(t, clientId, userInfo.GetClientId())
}

func TestGetMerchantId(t *testing.T) {
	merchantId := int64(789)
	userInfo := &UserInfo{merchantId: merchantId}
	assert.Equal(t, merchantId, userInfo.GetMerchantId())
}
