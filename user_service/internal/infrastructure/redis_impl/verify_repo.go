package redis_impl

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/pb/gen/user"
	"go_micro_service_api/user_service/internal/config"
	"go_micro_service_api/user_service/internal/domain/repository"
	"go_micro_service_api/user_service/internal/domain/vo"
	"time"
)

type VerifyRepo struct {
	cache db.Cache
}

var _ repository.VerifyRepo = (*VerifyRepo)(nil)

func NewVerifyRepo(cache db.Cache) repository.VerifyRepo {
	return &VerifyRepo{
		cache: cache,
	}
}

func (repo *VerifyRepo) RegisterVerification(ctx context.Context, session *vo.VerificationSession) *cus_err.CusError {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	cfg := config.GetConfig()

	// check notify lock
	// every 60 secs(vary) could request once
	err := repo.IsNotifyLockOn(ctx, session)
	if err != nil {
		return err
	}

	// check token lock
	// failed too many times will be locked for 3600 secs(vary)
	err = repo.IsPeriodLockOn(ctx, session)
	if err != nil {
		return err
	}

	// store code
	err = repo.cache.SetObject(ctx, session.GetCodeRedisKey(), session, time.Duration(cfg.VerificationTokenExpiry)*time.Second)
	if err != nil {
		return cus_err.New(cus_err.InternalServerError, "failed to set verification token", err)
	}

	// init error counts
	_, err = repo.cache.Get(ctx, session.GetErrorCountRedisKey())
	if err != nil {
		if err.Code().Int() == cus_err.ResourceNotFound {
			err = repo.cache.Set(ctx, session.GetErrorCountRedisKey(), "0", time.Duration(cfg.VerificationTokenCountPeriod)*time.Second)
			if err != nil {
				return cus_err.New(cus_err.InternalServerError, "failed to set verification token error count", err)
			}
		} else {
			return cus_err.New(cus_err.InternalServerError, "failed to get verification token error count", err)
		}
	}

	// init notify lock
	err = repo.cache.Set(ctx, session.GetNotifyLockRedisKey(), "true", time.Duration(cfg.VerificationTokenNotifyLockPeriod)*time.Second)
	if err != nil {
		return cus_err.New(cus_err.InternalServerError, "failed to set verification token notify lock", err)
	}

	return nil
}

func (repo *VerifyRepo) Verification(ctx context.Context, session *vo.VerificationSession) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	cfg := config.GetConfig()
	totalAttempts := cfg.VerificationTokenTotalAttempts

	// check token lock
	// failed too many times will be locked for 3600 secs(vary)
	err := repo.IsPeriodLockOn(ctx, session)
	if err != nil {
		return false, err
	}

	key := session.GetCodeRedisKey()

	// get old session from redis
	oldSession := new(vo.VerificationSession)
	err = repo.cache.GetObject(ctx, key, oldSession)
	if err != nil {
		if err.Code().Int() == cus_err.ResourceNotFound {
			return false, cus_err.New(cus_err.InvalidArgument, "verification token is expired", err)
		}
		return false, cus_err.New(cus_err.InvalidArgument, "failed to get verification token", err)
	}

	// verify code
	verified := oldSession.Verify(session.GetVerificationCode(), session.Token)

	// clean redis
	if verified {
		err := repo.cache.Delete(
			ctx,
			oldSession.GetCodeRedisKey(),
			oldSession.GetErrorCountRedisKey(),
			oldSession.GetNotifyLockRedisKey(),
			oldSession.GetLockRedisKey(),
		)
		if err != nil && err.Code().Int() != cus_err.ResourceNotFound {
			return false, cus_err.New(cus_err.InternalServerError, "failed to delete verification token", err)
		}
	} else {
		// incremented error count
		count, err := repo.cache.Incr(ctx, oldSession.GetErrorCountRedisKey())
		if err != nil {
			return false, cus_err.New(cus_err.InternalServerError, "failed to increment verification token error count", err)
		}

		if count >= int64(totalAttempts) {
			err := repo.cache.Set(ctx, oldSession.GetLockRedisKey(), "true", time.Duration(cfg.VerificationTokenLockPeriod)*time.Second)
			if err != nil {
				return false, cus_err.New(cus_err.InternalServerError, "failed to set verification token notify lock", err)
			}

			return false, NewInvalidVerificationError(
				"verification token is locked",
				int32(totalAttempts),
				int32(totalAttempts),
			)
		}

		return false, NewInvalidVerificationError(
			"verification code is invalid",
			int32(totalAttempts),
			int32(count),
		)
	}

	return verified, nil
}

func (repo *VerifyRepo) IsNotifyLockOn(ctx context.Context, session *vo.VerificationSession) *cus_err.CusError {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	cfg := config.GetConfig()

	lock, err := repo.cache.Get(ctx, session.GetNotifyLockRedisKey())
	if err != nil {
		if err.Code().Int() == cus_err.ResourceNotFound {
			return nil
		}
		return cus_err.New(cus_err.InternalServerError, "failed to get verification token notify lock", err)
	}
	// if register too many times in a period, return error
	if lock != "" {
		return NewInvalidVerificationError(
			"request verification token too fast",
			int32(cfg.VerificationTokenTotalAttempts),
			0,
		)
	}

	return nil
}

func (repo *VerifyRepo) IsPeriodLockOn(ctx context.Context, session *vo.VerificationSession) *cus_err.CusError {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	cfg := config.GetConfig()

	lock, err := repo.cache.Get(ctx, session.GetLockRedisKey())
	if err != nil {
		if err.Code().Int() == cus_err.ResourceNotFound {
			return nil
		}
		return cus_err.New(cus_err.InternalServerError, "failed to get verification token lock", err)
	}
	// if locked in a period, return error
	if lock != "" {
		return NewInvalidVerificationError(
			"verification token is locked",
			int32(cfg.VerificationTokenTotalAttempts),
			int32(cfg.VerificationTokenTotalAttempts),
		)
	}

	return nil
}

func NewInvalidVerificationError(message string, totalAttempts int32, errorCount int32) *cus_err.CusError {
	return cus_err.New(cus_err.InvalidVerificationCode, message).WithData(
		user.VerificationErrorResponse{
			TotalAttempts: totalAttempts,
			ErrorCount:    errorCount,
		},
	)
}
