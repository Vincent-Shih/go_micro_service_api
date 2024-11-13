package ent_impl

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/user_service/internal/domain/aggregate"
	"go_micro_service_api/user_service/internal/domain/repository"
	"go_micro_service_api/user_service/internal/domain/vo"
	"go_micro_service_api/user_service/internal/infrastructure/client/google"
	"go_micro_service_api/user_service/internal/infrastructure/client/meta"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/predicate"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/profile"
	"strings"

	"github.com/go-resty/resty/v2"
)

type UserRepo struct {
	db db.Database
}

var _ repository.UserRepo = (*UserRepo)(nil)

func NewUserRepo(db db.Database) repository.UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (repo *UserRepo) CreateProfile(ctx context.Context, u *aggregate.User) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	tx, ok := repo.db.GetTx(ctx).(*ent.Tx)
	if !ok {
		return nil, cus_err.New(cus_err.InternalServerError, "failed to get transaction", nil)
	}

	items := map[enum.Profile]string{
		enum.ProfileKey.Account:      u.Profile.Account,
		enum.ProfileKey.Email:        u.Profile.Email,
		enum.ProfileKey.CountryCode:  u.Profile.CountryCode,
		enum.ProfileKey.MobileNumber: u.Profile.MobileNumber,
	}
	ops := make([]*ent.ProfileCreate, 0)
	for k, v := range items {
		if strings.TrimSpace(v) != "" {
			ops = append(ops, tx.Profile.Create().SetKey(k.ID).SetUserID(int(u.ID)).SetValue(v))
		}
	}
	_, err := tx.Profile.CreateBulk(ops...).Save(ctx)
	if err != nil {
		return nil, cus_err.New(cus_err.InternalServerError, "failed to create profile", err)
	}

	return u, nil
}

func (repo *UserRepo) GetProfile(ctx context.Context, u *aggregate.User, keys []int) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := repo.db.GetClient(ctx).(*ent.Client)

	instances, err := client.Profile.Query().Where(profile.UserIDEQ(int(u.ID))).Where(profile.KeyIn(keys...)).All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			cusErr := cus_err.New(cus_err.ResourceNotFound, "profile not found", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}

		cusErr := cus_err.New(cus_err.InternalServerError, "failed to get profile", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	for _, instance := range instances {
		switch instance.Key {
		case enum.ProfileKey.Email.ID:
			u.Profile.Email = instance.Value
		case enum.ProfileKey.CountryCode.ID:
			u.Profile.CountryCode = instance.Value
		case enum.ProfileKey.MobileNumber.ID:
			u.Profile.MobileNumber = instance.Value
		}
	}

	return u, nil
}

func (repo *UserRepo) GetProfileFromOAuth(ctx context.Context, u *aggregate.User, session *vo.OAuthSession) (*aggregate.User, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := resty.New()
	switch session.Provider {
	case enum.OAuthProvider.Google:
		response, err := google.NewService(client).GetMe(ctx, session.AccessToken)
		if err != nil {
			return u, cus_err.New(cus_err.ThirdPartyError, "failed to get user info from oauth", err)
		}

		u.Profile.ThirdParties.Google.ID = response.ID
	case enum.OAuthProvider.Meta:
		response, err := meta.NewService(client).GetMe(ctx, session.AccessToken)
		if err != nil {
			return u, cus_err.New(cus_err.ThirdPartyError, "failed to get user info from oauth", err)
		}

		u.Profile.ThirdParties.Meta.ID = response.ID
	case enum.OAuthProvider.Twitter:
		// x.NewService().GetMe(ctx, req.AccessToken)
	}

	return u, nil
}

func (repo *UserRepo) CheckMobileExistence(ctx context.Context, countryCode string, mobileNumber string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := repo.db.GetClient(ctx).(*ent.Client)

	conds := []predicate.Profile{
		profile.And(
			profile.KeyEQ(enum.ProfileKey.CountryCode.ID),
			profile.ValueEQ(countryCode),
		),
		profile.And(
			profile.KeyEQ(enum.ProfileKey.MobileNumber.ID),
			profile.ValueEQ(mobileNumber),
		),
	}

	counts, err := client.Profile.Query().
		Where(profile.Or(conds...)).
		Count(ctx)
	if err != nil {
		return false, cus_err.New(cus_err.InternalServerError, "failed to query profile", err)
	}

	return counts == len(conds), nil
}

func (repo *UserRepo) CheckEmailExistence(ctx context.Context, email string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := repo.db.GetClient(ctx).(*ent.Client)

	exist, err := client.Profile.Query().
		Where(
			profile.Key(enum.ProfileKey.Email.ID),
			profile.Value(email),
		).Exist(ctx)
	if err != nil {
		return true, cus_err.New(cus_err.InternalServerError, "failed to query error", err)
	}
	return exist, nil
}

func (repo *UserRepo) IsAccountExist(ctx context.Context, account string) (bool, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := repo.db.GetClient(ctx).(*ent.Client)

	exist, err := client.Profile.Query().
		Where(
			profile.Key(enum.LoginTypes.Account.Id),
			profile.Value(account),
		).Exist(ctx)

	if err != nil {
		return false, cus_err.New(cus_err.InternalServerError, "failed to query error", err)
	}
	return exist, nil
}

func (repo *UserRepo) GetUserIdByProfile(ctx context.Context, mapping map[int]string) (int, *cus_err.CusError) {
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	client := repo.db.GetClient(ctx).(*ent.Client)

	predicates := []predicate.Profile{}
	for k, v := range mapping {
		predicates = append(predicates, profile.And(profile.KeyEQ(k), profile.ValueEQ(v)))
	}
	instance, err := client.Profile.Query().Where(predicates...).Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return 0, cus_err.New(cus_err.ResourceNotFound, "profile not found", err)
		}
		if ent.IsNotSingular(err) {
			return 0, cus_err.New(cus_err.AccountPasswordError, "multiple profiles found", err)
		}
		return 0, cus_err.New(cus_err.InternalServerError, "failed to get profile", err)
	}
	return instance.UserID, nil
}
