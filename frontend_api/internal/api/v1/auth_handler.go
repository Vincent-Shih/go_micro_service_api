package v1_handler

import (
	"go_micro_service_api/frontend_api/internal/infrastructure/grpc_client"
	"go_micro_service_api/frontend_api/internal/model/request"
	"go_micro_service_api/frontend_api/internal/model/response"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/pb/gen/auth"
	"go_micro_service_api/pkg/responder"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authGrpc *grpc_client.AuthClient
	userGrpc *grpc_client.UserClient
}

func NewAuthHandler(authGrpc *grpc_client.AuthClient, userGrpc *grpc_client.UserClient) *AuthHandler {
	return &AuthHandler{
		authGrpc: authGrpc,
		userGrpc: userGrpc,
	}
}

// ShowAccount godoc
// @Summary      客戶端驗證，並且取得 JWT 簽名
// @Description  這個 api 主要是要讓客戶端驗證，並且取得 JWT 簽名，如果拿到簽名的話，在後續就可以根據簽名來存取 api，client_id 當有營運單位提出對接時，由大後台產生並提供給前端
// @Tags         Auth
// @Produce      json
// @Version      1.0
// @Param client_id header string true "Client ID"
// @Success      200  	{object}    response.Response{data=response.TokenResponse}
// @Failure      400  	{object}  	response.Response
// @Failure      404  	{object}  	response.Response
// @Failure      500  	{object}  	response.Response
// @Router       /v1/auth/ [get]
func (a *AuthHandler) ClientAuth(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get access token from Authorization header
	// Expected format: "Bearer <token>"
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Extract the token from the Authorization header
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) == 2 {
			accessToken := splitToken[1]
			// Validate the access token
			_, validErr := a.authGrpc.ValidToken(ctx, accessToken)
			if validErr == nil {
				responder.Ok(nil).WithContext(c)
				return
			}
		}
	}

	// Get the client id from Header
	val := c.GetHeader("client_id")
	if val == "" {
		cusErr := cus_err.New(cus_err.AccountPasswordError, "Client id not found in header", nil)
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}
	// Convert the client id to int
	clientId, err := strconv.Atoi(val)
	if err != nil {
		cusErr := cus_err.New(cus_err.AccountPasswordError, "Invalid client id", err)
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	// Call the auth grpc
	res, cusErr := a.authGrpc.ClientAuth(ctx, int64(clientId))
	if cusErr != nil {
		responder.Error(cusErr).WithContext(c)
		return
	}

	responder.Ok(&response.TokenResponse{AccessToken: res.AccessToken}).WithContext(c)
}

// Login login
// @Summary      登入
// @Description  這個 api 主要是要讓客戶端登入，登入成功後會回傳 JWT 簽名，如果拿到簽名的話，在後續就可以根據簽名來存取 api
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Version      1.0
// @Security Bearer
// @Param body body request.LoginRequest true "Login Request"
// @Success      200  	{object}	response.Response{data=response.LoignPassResponse}
// @Failure      400  	{object}  	response.Response
// @Failure      401  	{object}  	response.Response{data=response.LoginAnomalousResponse}
// @Failure      404  	{object}  	response.Response
// @Failure      500  	{object}  	response.Response
// @Router       /v1/users/login/ [post]
func (a *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get access token from Authorization header
	var accessToken string
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) == 2 {
			accessToken = splitToken[1]
		}
	}
	if accessToken == "" {
		cusErr := cus_err.New(cus_err.MissingAccessToken, "Access token not found in header", nil)
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	// Get the login request from the body
	var loginRequest request.LoginRequest

	if err := c.ShouldBindBodyWithJSON(&loginRequest); err != nil {
		// New a Cuserr with InvalidArgument error code and log it berfore returning
		cusErr := cus_err.New(cus_err.AccountPasswordError, "Invalid request", err)
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	// Get user agent and IP address from th header
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	userInfo, cusErr := a.userGrpc.GetLoginUserInfo(ctx, loginRequest)

	// 如有搜尋的錯誤代表搜尋不到結果
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		responder.Error(cus_err.New(cus_err.AccountError, "Account not found")).WithContext(c)
		return
	}

	// Authenticate the user
	loginInfo := &auth.LoginRequest{
		UserAgent:   userAgent,
		Ip:          ip,
		UserId:      userInfo.UserId,
		AccessToken: accessToken,
		Password:    loginRequest.Password,
	}
	res, cusErr := a.authGrpc.Login(ctx, loginInfo)
	if cusErr != nil {
		// 直接使用 cusErr，因為錯誤資訊已經在 client 層處理好了
		switch cusErr.Code().Int() {
		case cus_err.AccountPasswordError:
			// 處理密碼錯誤
			cus_otel.Error(ctx, cusErr.Error())
			responder.Error(cusErr).WithContext(c)
			return

		case cus_err.UnusualLogin:
			// 處理異常登入
			cusErr.WithData(response.LoginAnomalousResponse{
				Email:        userInfo.Email,
				CountryCode:  userInfo.CountryCode,
				MobileNumber: userInfo.MobileNumber,
			})
			responder.Error(cusErr).WithContext(c)
			return

		default:
			// 處理其他可能的錯誤
			responder.Error(cusErr).WithContext(c)
			return

		}
	}

	responder.Ok(&response.TokenResponse{AccessToken: res.AccessToken}).WithContext(c)
}
