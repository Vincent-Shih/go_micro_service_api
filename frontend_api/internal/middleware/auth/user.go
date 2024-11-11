package auth

import (
	"go_micro_service_api/pkg/enum"

	"github.com/gin-gonic/gin"
)

// UserInfo represents the authenticated user's information for the current request scope.
// This struct is populated during the authentication process and stored in the request context.
// It contains essential details about the user making the API call, including permissions and identifiers.
//
// Usage:
//   - This struct is created and populated by the authentication middleware.
//   - After successful authentication, it is stored in the request's context.
//   - Subsequent middleware and request handlers can retrieve and use this information
//     to make authorization decisions or personalize the response.
//
// Note: The UserInfo instance is specific to each request and should not be stored
// or used beyond the lifetime of the request.
type UserInfo struct {
	permissions []enum.Permission // List of permissions granted to the user
	userAccount *string           // The user's account may be null
	userId      *int64            // The user's id
	clientId    int64             // The client's id
	merchantId  int64             // Whether in the frontend or backend, the merchant  under the same business entity should be consistent.
}

// NewUserInfo creates a new UserInfo instance with the provided details.
func NewUserInfo(
	permissions []enum.Permission,
	clientId int64,
	merchantId int64,
	userAccount *string,
	userId *int64,
) *UserInfo {
	return &UserInfo{
		permissions: permissions,
		clientId:    clientId,
		merchantId:  merchantId,
		userAccount: userAccount,
		userId:      userId,
	}
}

// GetUserInfo retrieves the user information from the Gin context.
func GetUserInfo(c *gin.Context) (userInfo *UserInfo, ok bool) {
	u, ok := c.Get(userInfoKey)
	if !ok {
		return nil, false
	}

	// Check if the user information is of the correct type
	if _, ok := u.(*UserInfo); !ok {
		return nil, false
	}

	return u.(*UserInfo), true
}

// HasPermission checks if the user has all the specified permissions.
//
// This method verifies whether the user possesses all the permissions
// provided as arguments. It returns true only if the user has every
// single one of the specified permissions.
//
// Parameters:
//   - needs: A variadic parameter of type Permission, representing
//     the permissions to check for.
//
// Returns:
//   - bool: true if the user has all the specified permissions,
//     false otherwise.
func (u *UserInfo) HasPermission(needs ...enum.Permission) bool {
	// Create a map for faster lookup of user's permissions
	userPerms := make(map[enum.Permission]bool)
	for _, perm := range u.permissions {
		userPerms[perm] = true
	}

	// Check if the user has all needed permissions
	for _, need := range needs {
		if !userPerms[need] {
			return false // User is missing this permission
		}
	}

	return true // User has all needed permissions
}

// GetPermissions returns the user's permissions.
func (u *UserInfo) GetPermissions() []enum.Permission {
	return append([]enum.Permission(nil), u.permissions...)
}

// GetUserId returns the user's id.
// if the user's id is not set, it returns 0 and false.
func (u *UserInfo) GetUserId() (userId int64, ok bool) {
	if u.userId == nil {
		return 0, false
	}
	return *u.userId, true
}

// GetUserAccount returns the user's account.
// when the user's account is not set , it returns an empty string and false.
func (u *UserInfo) GetUserAccount() (userAccount string, ok bool) {
	if u.userAccount == nil {
		return "", false
	}
	return *u.userAccount, true
}

// GetClientId returns the client's id.
func (u *UserInfo) GetClientId() int64 {
	return u.clientId
}

// GetMerchantId returns the merchant's id.
func (u *UserInfo) GetMerchantId() int64 {
	return u.merchantId
}
