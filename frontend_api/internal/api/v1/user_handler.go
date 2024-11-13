package v1_handler

import (
	"go_micro_service_api/frontend_api/internal/infrastructure/grpc_client"
	auth_middleware "go_micro_service_api/frontend_api/internal/middleware/auth"
	"go_micro_service_api/frontend_api/internal/model/request"
	"go_micro_service_api/frontend_api/internal/model/response"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/helper"
	"go_micro_service_api/pkg/pb/gen/auth"
	"go_micro_service_api/pkg/pb/gen/user"
	"go_micro_service_api/pkg/responder"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authGrpc        *grpc_client.AuthClient
	userGrpc        *grpc_client.UserClient
	snowFlakeHelper *helper.Snowflake
}

func NewUserHandler(authGrpc *grpc_client.AuthClient, userGrpc *grpc_client.UserClient, snowFlakeHelper *helper.Snowflake) *UserHandler {
	return &UserHandler{
		authGrpc:        authGrpc,
		userGrpc:        userGrpc,
		snowFlakeHelper: snowFlakeHelper,
	}
}

// @Summary Create User
// @Description Create User
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body request.RegisterRequest true "Register Request"
// @Success 200 {object} response.Response{data=response.RegisterResponse}
// @Failure 400 {object} response.Response{data=response.VerificationErrorResponse}
// @Failure 401 {object} response.Response
// @Router /v1/users/ [post]
func (u *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	userInfo, ok := auth_middleware.GetUserInfo(c)
	if !ok {
		cusErr := cus_err.New(cus_err.Unauthorized, "Unauthenticated")
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	// body validation
	var request request.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		cusErr := cus_err.New(cus_err.AccountPasswordError, "Invalid request", err)
		cus_otel.Warn(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	// verification
	res, err := u.userGrpc.Verification(ctx, &user.VerificationRequest{
		Email:                  request.Email,
		CountryCode:            request.CountryCode,
		MobileNumber:           request.MobileNumber,
		VerificationCodePrefix: request.VerificationCodePrefix,
		VerificationCode:       request.VerificationCode,
		VerificationCodeToken:  request.VerificationCodeToken,
	})

	// WARN: err should contain user.VerificationErrorResponse
	if err != nil || !res.GetResult() {
		responder.Error(err).WithContext(c)
		return
	}

	userId := u.snowFlakeHelper.NextID()

	// TODO: XA

	// create user identity on auth service
	err = u.authGrpc.CreateUser(ctx, &auth.CreateUserRequest{
		ClientId: userInfo.GetClientId(),
		Id:       userId,
		Account:  request.Account,
		Password: request.Password,
		Status:   int32(enum.UserStatusType.Active),
	})
	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}

	// create user profile on user service
	_, err = u.userGrpc.CreateProfile(ctx, &user.CreateProfileRequest{
		Id:           userId,
		Account:      request.Account,
		Email:        request.Email,
		CountryCode:  request.CountryCode,
		MobileNumber: request.MobileNumber,
	})
	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}

	// loginRes, err := u.authGrpc.Login(ctx, &auth.LoginRequest{
	// 	UserId:    userId,
	// 	Password:  "",
	// 	UserAgent: "",
	// 	Ip:        "",
	// })

	responder.Ok(response.RegisterResponse{
		// AccessToken:  loginRes.GetAccessToken(),
		Account:      request.Account,
		Email:        request.Email,
		CountryCode:  request.CountryCode,
		MobileNumber: request.MobileNumber,
	}).WithContext(c)
}

// @Summary 檢查email / mobile number+ country code / account是否存在
// @Description Check User Existence
// @Tags User
// @Produce json
// @Security Bearer
// @Param account query string false "Account"
// @Param email query string false "Email"
// @Param mobileNumber query string false "Mobile Number"
// @Param countryCode query string false "Country Code"
// @Success 200 {object} response.Response{data=response.ExistenceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /v1/users/existence [get]
func (u *UserHandler) CheckUserExistence(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	var request request.CheckUserExistenceRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		cusErr := cus_err.New(cus_err.AccountPasswordError, "account, mobileNumber or email is required", err)
		cus_otel.Error(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	var exists bool
	var cusErr *cus_err.CusError
	switch {
	case request.Account != "":
		exists, cusErr = u.authGrpc.CheckAccountExistence(ctx, request.Account)
	case request.MobileNumber != "" && request.CountryCode != "":
		exists, cusErr = u.userGrpc.CheckMobileExistence(ctx, request.MobileNumber, request.CountryCode)
	case request.Email != "":
		exists, cusErr = u.userGrpc.CheckEmailExistence(ctx, request.Email)
	}

	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		responder.Error(cusErr).WithContext(c)
		return
	}

	responder.Ok(response.ExistenceResponse{
		Exists: exists,
	}).WithContext(c)
}
