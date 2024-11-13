package auth

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/enum"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setUserInfo(u *UserInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(userInfoKey, u)
		c.Next()
	}
}

// errorHandler is a simple middleware that handles errors for this test.
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		cusErr, ok := err.Err.(*cus_err.CusError)
		if ok {
			c.Status(cusErr.HttpCode())
			return
		} else {
			c.Status(http.StatusInternalServerError)
			return
		}

	}
}

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Close() error {
	return nil
}

func (m *MockAuthClient) ValidToken(ctx context.Context, token string) (*UserInfo, *cus_err.CusError) {
	args := m.Called(ctx, token)
	return args.Get(0).(*UserInfo), args.Get(1).(*cus_err.CusError)
}

func TestGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userInfo       UserInfo
		options        []PermOption
		expectedStatus int
	}{
		{
			name: "UserHasAllPermissions",
			userInfo: UserInfo{
				permissions: []enum.Permission{enum.PermissionType.Withdraw, enum.PermissionType.Deposit},
			},
			options: []PermOption{
				WithPerms(enum.PermissionType.Deposit, enum.PermissionType.Withdraw),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "UserMissingPermission",
			userInfo: UserInfo{
				permissions: []enum.Permission{enum.PermissionType.Deposit},
			},
			options: []PermOption{
				WithPerms(enum.PermissionType.Deposit, enum.PermissionType.Withdraw),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "NoSetPermOptions",
			userInfo: UserInfo{
				permissions: []enum.Permission{enum.PermissionType.Deposit},
			},
			options:        []PermOption{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			r.Use(errorHandler())
			r.Use(setUserInfo(&tt.userInfo))
			r.GET("/test", Guard(tt.options...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345" // Set IP address
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})

	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                 string
		path                 string
		token                string
		bypassPaths          []string
		filters              []Filter
		mockResponse         *UserInfo
		mockError            *cus_err.CusError
		expectValidTokenCall bool
		expectedStatus       int
	}{
		{
			name:                 "Bypass path",
			path:                 "/api/v1/public/status",
			bypassPaths:          []string{"/api/v1/public/*"},
			expectValidTokenCall: false,
			expectedStatus:       http.StatusOK,
		},
		{
			name: "ByPass with filter",
			path: "/api/v1/public/status",
			filters: []Filter{
				func(c *gin.Context) bool {
					testFlag := c.GetHeader("Test")
					return testFlag == "true"
				},
			},
			expectValidTokenCall: false,
			expectedStatus:       http.StatusOK,
		},
		{
			name:                 "Valid token",
			path:                 "/api/v1/protected",
			token:                "valid_token",
			expectValidTokenCall: true,
			mockResponse: &UserInfo{
				permissions: []enum.Permission{enum.PermissionType.Deposit, enum.PermissionType.Withdraw},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                 "Invalid token",
			path:                 "/api/v1/protected",
			token:                "invalid_token",
			mockError:            cus_err.New(cus_err.Unauthorized, "invalid token"),
			expectValidTokenCall: true,
			expectedStatus:       http.StatusUnauthorized,
		},
		{
			name:                 "Missing token",
			path:                 "/api/v1/protected",
			mockError:            cus_err.New(cus_err.MissingAccessToken, "missing access token"),
			expectValidTokenCall: false,
			expectedStatus:       http.StatusUnauthorized,
		},
		{
			name:                 "Not set user info",
			path:                 "/api/v1/protected",
			token:                "valid_token",
			mockResponse:         nil,
			mockError:            nil,
			expectValidTokenCall: true,
			expectedStatus:       http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockAuthClient)
			// Set the expected call to the ValidToken method
			if tt.expectValidTokenCall {
				mockClient.On("ValidToken", mock.Anything, tt.token).Return(tt.mockResponse, tt.mockError)
			}

			// Create a new gin engine for each test
			r := gin.New()
			r.Use(errorHandler())
			r.Use(AuthMiddleware(mockClient, ByPassPath(tt.bypassPaths...), WithFilter(tt.filters...)))
			r.GET(tt.path, func(c *gin.Context) {
				// Check if UserInfo was set correctly in the context
				if tt.mockResponse != nil {
					userInfo, exists := c.Get(userInfoKey)
					assert.True(t, exists)
					assert.Equal(t, tt.mockResponse, userInfo)
				}

				c.Status(http.StatusOK)
			})

			// Perform the request
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), "GET", tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("Test", "true")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestIsBypassPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		bypass   []string
		expected bool
	}{
		{"Exact match", "/api/v1/user", []string{"/api/v1/user"}, true},
		{"Wildcard match", "/api/v1/public/status", []string{"/api/v1/public/*"}, true},
		{"No match", "/api/v1/private", []string{"/api/v1/public/*"}, false},
		{"wildcard pattern *any", "/api/v1/swagger/*any", []string{"/api/v1/swagger/*"}, true},
		{"Multiple patterns", "/api/v1/user", []string{"/api/v1/public/*", "/api/v1/user"}, true},
		{"wildcard in the middle", "/api/v1/user/profile", []string{"/api/*/user/profile"}, true},
		{"wildcard in the middle layer not match", "/api/v1/user/profile", []string{"/api/v1/*/*/profile"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBypassPath(tt.path, tt.bypass)
			assert.Equal(t, tt.expected, result)
		})
	}
}
