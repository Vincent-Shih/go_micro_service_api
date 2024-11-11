package v1_handler

import (
	"go_micro_service_api/frontend_api/internal/infrastructure/grpc_client"
	"go_micro_service_api/frontend_api/internal/model/request"
	"go_micro_service_api/frontend_api/internal/model/response"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/user"
	"go_micro_service_api/pkg/responder"

	"github.com/gin-gonic/gin"
)

type VerifyHandler struct {
	userGrpc *grpc_client.UserClient
}

func NewVerifyHandler(userGrpc *grpc_client.UserClient) *VerifyHandler {
	return &VerifyHandler{
		userGrpc: userGrpc,
	}
}

// @Summary 申請驗證碼
// @Description Register Verification
// @Tags User
// @Produce json
// @Security Bearer
// @Param type query string false "Type" Enums(forgotPwd,unusualLogin)
// @Param email query string false "Email"
// @Param countryCode query string false "Country Code"
// @Param mobileNumber query string false "Mobile Number"
// @Success 200 {object} response.Response{data=response.RegisterVerificationResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /v1/users/verificationCode/ [get]
func (v *VerifyHandler) RegisterVerification(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// query validation
	var request request.RegisterVerificationRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		kgsErr := cus_err.New(cus_err.AccountPasswordError, "Invalid request", err)
		cus_otel.Warn(ctx, kgsErr.Error())
		responder.Error(kgsErr).WithContext(c)
		return
	}

	res, err := v.userGrpc.RegisterVerification(ctx, &user.RegisterVerificationRequest{
		Type:         request.Type,
		Email:        request.Email,
		CountryCode:  request.CountryCode,
		MobileNumber: request.MobileNumber,
	})
	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}

	// TODO: if failed to send code then rollback
	// TODO: send verification code

	response := &response.RegisterVerificationResponse{
		VerificationCodePrefix: res.VerificationCodePrefix,
		VerificationCodeToken:  res.VerificationCodeToken,
	}
	responder.Ok(response).WithContext(c)
}

// @Summary 驗證
// @Description Verification
// @Tags User
// @Produce json
// @Security Bearer
// @Param body body request.VerificationRequest true "Verification Request"
// @Success 200 {object} response.Response{data=response.VerificationResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /v1/users/verification/ [post]
func (v *VerifyHandler) Verification(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// body validation
	var req request.VerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		kgsErr := cus_err.New(cus_err.InvalidArgument, "Invalid request", err)
		cus_otel.Warn(ctx, kgsErr.Error())
		responder.Error(kgsErr).WithContext(c)
		return
	}

	res, err := v.userGrpc.Verification(ctx, &user.VerificationRequest{
		Type:                   req.Type,
		Email:                  req.Email,
		CountryCode:            req.CountryCode,
		MobileNumber:           req.MobileNumber,
		VerificationCodePrefix: req.VerificationCodePrefix,
		VerificationCode:       req.VerificationCode,
		VerificationCodeToken:  req.VerificationCodeToken,
	})
	// WARN: err should contain user.VerificationErrorResponse
	if err != nil || !res.GetResult() {
		responder.Error(err).WithContext(c)
		return
	}

	getUserInfoReq := request.LoginRequest{
		Email:        req.Email,
		CountryCode:  req.CountryCode,
		MobileNumber: req.MobileNumber,
	}
	if req.Email != "" {
		getUserInfoReq.LoginType = enum.LoginTypes.Email.String
	}
	if req.CountryCode != "" && req.MobileNumber != "" {
		getUserInfoReq.LoginType = enum.LoginTypes.MobileNumber.String
	}
	userInfo, err := v.userGrpc.GetLoginUserInfo(ctx, getUserInfoReq)
	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}

	response := &response.VerificationResponse{
		Account:      userInfo.Account,
		Email:        userInfo.Email,
		CountryCode:  userInfo.CountryCode,
		MobileNumber: userInfo.MobileNumber,
	}

	// login if unusual
	// TODO: change to notify event enum
	if req.Type == "unusualLogin" {

	}

	responder.Ok(response).WithContext(c)
}
