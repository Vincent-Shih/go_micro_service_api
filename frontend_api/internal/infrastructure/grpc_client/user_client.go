package grpc_client

import (
	"context"
	"go_micro_service_api/frontend_api/internal/config"
	"go_micro_service_api/frontend_api/internal/model/request"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	otelgrpc "go_micro_service_api/pkg/cus_otel/grpc"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/pkg/pb/gen/user"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn             *grpc.ClientConn
	userGrpcClient   user.UserServiceClient
	verifyGrpcClient user.VerifyServiceClient
}

func NewUserClient(cfg *config.Config) (*UserClient, error) {
	// Get address from config
	gprcAddr := cfg.UserUrl

	// New grpc client with own tracing middleware
	conn, err := grpc.NewClient(
		gprcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.TracingMiddleware(otelgrpc.RoleClient)),
	)
	if err != nil {
		return &UserClient{}, err
	}
	userGrpc := user.NewUserServiceClient(conn)
	verifyGrpc := user.NewVerifyServiceClient(conn)

	return &UserClient{
		conn:             conn,
		userGrpcClient:   userGrpc,
		verifyGrpcClient: verifyGrpc,
	}, nil
}

// Close terminates the gRPC connection associated with the UserClient.
// It should be called when the client is no longer needed to free up resources.
//
// Returns:
//   - error: An error if closing the connection fails, nil otherwise.
func (c *UserClient) Close() error {
	return c.conn.Close()
}

// CreateProfile creates a new user profile with the provided profile information.
func (a *UserClient) CreateProfile(ctx context.Context, req *user.CreateProfileRequest) (*user.CreateProfileResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	if req.Id <= 0 {
		err := cus_err.New(cus_err.InvalidArgument, "user ID is required", nil)
		cus_otel.Error(ctx, err.Error())
		return &user.CreateProfileResponse{}, err
	}

	res, grpcErr := a.userGrpcClient.CreateProfile(ctx, req)
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return &user.CreateProfileResponse{}, err
		}
		err := cus_err.New(cus_err.InternalServerError, "can't found the kgsErr from grpcErr", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return &user.CreateProfileResponse{}, err
	}

	return res, nil
}

// FindProfile finds a user profile by the provided user ID.
func (a *UserClient) FindProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	if req.Id <= 0 {
		err := cus_err.New(cus_err.InvalidArgument, "user ID is required", nil)
		cus_otel.Error(ctx, err.Error())
		return &user.GetProfileResponse{}, err
	}

	res, grpcErr := a.userGrpcClient.GetProfile(ctx, req)
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return &user.GetProfileResponse{}, err
		}
		err := cus_err.New(cus_err.InternalServerError, "can't found the kgsErr from grpcErr", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return &user.GetProfileResponse{}, err
	}

	return res, nil
}

// RegisterVerification register a verification session in redis.
func (a *UserClient) RegisterVerification(ctx context.Context, req *user.RegisterVerificationRequest) (*user.RegisterVerificationResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	res, grpcErr := a.verifyGrpcClient.RegisterVerification(ctx, req)
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return &user.RegisterVerificationResponse{}, err
		}
		err := cus_err.New(cus_err.InternalServerError, "can't found the kgsErr from grpcErr", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return &user.RegisterVerificationResponse{}, err
	}

	return res, nil
}

// Verification verifies the user's verification code.
func (a *UserClient) Verification(ctx context.Context, req *user.VerificationRequest) (*user.VerificationResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	if req.VerificationCodePrefix == "" || req.VerificationCode == "" {
		err := cus_err.New(cus_err.InvalidArgument, "verification code, verification code prefix are required", nil)
		cus_otel.Error(ctx, err.Error())
		return &user.VerificationResponse{}, err
	}

	res, grpcErr := a.verifyGrpcClient.Verification(ctx, req)
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return &user.VerificationResponse{}, err
		}
		err := cus_err.New(cus_err.InternalServerError, "can't found the kgsErr from grpcErr", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return &user.VerificationResponse{}, err
	}

	return res, nil
}

// CheckMobileExistence checks if the mobile number exists in the database.
func (u *UserClient) CheckMobileExistence(ctx context.Context, mobileNumber string, countryCode string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	validate := validator.New()
	errs := validate.ValidateMap(map[string]interface{}{
		"mobileNumber": mobileNumber,
		"countryCode":  countryCode,
	}, map[string]interface{}{
		"mobileNumber": "required,numeric",
		"countryCode":  "required,numeric",
	})
	if len(errs) > 0 {
		err := cus_err.New(
			cus_err.InvalidArgument,
			"mobile number & country code is invalid",
			errs["mobileNumber"].(error),
			errs["countryCode"].(error),
		)
		cus_otel.Error(ctx, err.Error())
		return false, err
	}

	res, grpcErr := u.userGrpcClient.CheckMobileExistence(ctx, &user.MobileExistenceRequest{
		MobileNumber: mobileNumber,
		CountryCode:  countryCode,
	})
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return true, err
		}

		err := cus_err.New(cus_err.ResponseNotFound, "grpc error", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return true, err
	}
	return res.Exist, nil
}

// CheckEmailExistence checks if the email exists in the database.
func (u *UserClient) CheckEmailExistence(ctx context.Context, email string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	validate := validator.New()
	err := validate.Var(email, "email")
	if err != nil {
		err := cus_err.New(cus_err.InvalidArgument, "email is invalid", err)
		cus_otel.Error(ctx, err.Error())
		return false, err
	}

	res, grpcErr := u.userGrpcClient.CheckEmailExistence(ctx, &user.EmailExistenceRequest{
		Email: email,
	})
	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return true, err
		}

		err := cus_err.New(cus_err.ResponseNotFound, "grpc error", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return true, err
	}
	return res.Exist, nil
}

// loginGetUserList retrieves the user list by the provided identifier.
func (u *UserClient) GetLoginUserInfo(ctx context.Context, payload request.LoginRequest) (*user.GetLoginUserInfoResponse, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	_, err := enum.LoginTypeFromString(payload.LoginType)
	if err != nil {
		return &user.GetLoginUserInfoResponse{}, cus_err.New(cus_err.InvalidArgument, "invalid login type")
	}

	req := &user.GetLoginUserInfoRequest{
		LoginType:    payload.LoginType,
		Email:        payload.Email,
		Account:      payload.Account,
		CountryCode:  payload.CountryCode,
		MobileNumber: payload.MobileNumber,
	}

	res, grpcErr := u.userGrpcClient.GetLoginUserInfo(ctx, req)

	if grpcErr != nil {
		if err, ok := cus_err.FromGrpcErr(grpcErr); ok {
			cus_otel.Error(ctx, err.Error())
			return &user.GetLoginUserInfoResponse{}, err
		}

		err := cus_err.New(cus_err.ResponseNotFound, "grpc error", grpcErr)
		cus_otel.Error(ctx, err.Error())
		return &user.GetLoginUserInfoResponse{}, err
	}
	return res, nil
}
