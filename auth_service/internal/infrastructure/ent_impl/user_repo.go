package ent_impl

import (
	"context"
	"go_micro_service_api/auth_service/internal/domain/aggregate"
	"go_micro_service_api/auth_service/internal/domain/entity"
	"go_micro_service_api/auth_service/internal/domain/repository"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent/loginrecord"
	"go_micro_service_api/auth_service/internal/infrastructure/ent_impl/ent/user"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"strings"
)

type UserRepoImpl struct {
	db db.Database
}

var _ repository.UserRepo = (*UserRepoImpl)(nil)

func NewUserRepoImpl(db db.Database) *UserRepoImpl {
	return &UserRepoImpl{
		db: db,
	}
}

func (u *UserRepoImpl) Find(ctx context.Context, id int64) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get client with transaction if exists
	var client *ent.Client
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if ok {
		client = tx.Client()
	} else {
		client = u.db.GetConn(ctx).(*ent.Client)
	}

	// Find user
	entUser, err := client.User.Query().
		Where(user.ID(id)).
		WithAuthClients().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			cusErr := cus_err.New(cus_err.ResourceNotFound, "user not found", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		cusErr := cus_err.New(cus_err.InternalServerError, "find user failed", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to enum.UserStatus
	status, cusErr := enum.UserStatusFromInt(entUser.Status)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to aggregate.User
	user := &aggregate.User{
		Id:                entUser.ID,
		Password:          entUser.Password,
		PasswordFailTimes: entUser.PasswordFailTimes,
		Status:            status,
	}
	setUserLoader(u.db, user)

	return user, nil
}

func (u *UserRepoImpl) Create(ctx context.Context, clientId int64, user *aggregate.User) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get Tx from context
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if !ok {
		err := cus_err.New(cus_err.InternalServerError, "get tx from context failed")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Validate parameters
	cusErr := u.validateParameters(user)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Check the client is exist
	_, err := tx.AuthClient.Get(ctx, clientId)
	if err != nil {
		cusErr := cus_err.New(cus_err.ResourceNotFound, "client not found", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Create user
	entUser, err := tx.User.Create().
		SetID(user.Id).
		SetAccount(user.Account).
		SetPassword(user.Password).
		SetPasswordFailTimes(user.PasswordFailTimes).
		SetStatus(user.Status.Int()).
		SetAuthClientsID(clientId).
		Save(ctx)
	if ent.IsConstraintError(err) && strings.Contains(err.Error(), "duplicate") {
		cusErr := cus_err.New(cus_err.ResourceIsExist, "user already exists", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "create user failed", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to enum.UserStatus
	status, cusErr := enum.UserStatusFromInt(entUser.Status)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to aggregate.User
	newUser := &aggregate.User{
		Id:                entUser.ID,
		Account:           entUser.Account,
		Password:          entUser.Password,
		PasswordFailTimes: entUser.PasswordFailTimes,
		Status:            status,
	}
	setUserLoader(u.db, newUser)

	return newUser, nil
}

func (u *UserRepoImpl) Update(ctx context.Context, user *aggregate.User) (*aggregate.User, *cus_err.CusError) {
	// Start trace

	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get Tx from context
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if !ok {
		err := cus_err.New(cus_err.InternalServerError, "get tx from context failed")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	cusErr := u.validateParameters(user)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Update user
	entUser, err := tx.User.UpdateOneID(user.Id).
		SetPassword(user.Password).
		SetPasswordFailTimes(user.PasswordFailTimes).
		SetStatus(user.Status.Int()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			cusErr := cus_err.New(cus_err.ResourceNotFound, "user not found", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		cusErr := cus_err.New(cus_err.InternalServerError, "update user failed", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to enum.UserStatus
	status, cusErr := enum.UserStatusFromInt(entUser.Status)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to aggregate.User
	updatedUser := &aggregate.User{
		Id:                entUser.ID,
		Account:           entUser.Account,
		Password:          entUser.Password,
		PasswordFailTimes: entUser.PasswordFailTimes,
		Status:            status,
	}
	setUserLoader(u.db, updatedUser)

	return updatedUser, nil
}

func (u *UserRepoImpl) AddLoginRecord(ctx context.Context, userId int64, loginRecord *entity.LoginRecord) (*entity.LoginRecord, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get Tx from context
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if !ok {
		err := cus_err.New(cus_err.InternalServerError, "get tx from context failed")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Create login record
	entLoginRecord, err := tx.LoginRecord.Create().
		SetBrowser(loginRecord.Browser).
		SetBrowserVer(loginRecord.BrowserVer).
		SetIP(loginRecord.Ip).
		SetOs(loginRecord.Os).
		SetPlatform(loginRecord.Platform).
		SetCountry(loginRecord.Country).
		SetCountryCode(loginRecord.CountryCode).
		SetCity(loginRecord.City).
		SetAsp(loginRecord.Asp).
		SetIsMobile(loginRecord.IsMobile).
		SetIsSuccess(loginRecord.IsSuccess).
		SetErrMessage(loginRecord.ErrMessage).
		SetUsersID(userId).
		Save(ctx)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "create login record failed", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to entity.LoginRecord
	return &entity.LoginRecord{
		Id:          entLoginRecord.ID,
		Browser:     entLoginRecord.Browser,
		BrowserVer:  entLoginRecord.BrowserVer,
		Ip:          entLoginRecord.IP,
		Os:          entLoginRecord.Os,
		Platform:    entLoginRecord.Platform,
		Country:     entLoginRecord.Country,
		CountryCode: entLoginRecord.CountryCode,
		City:        entLoginRecord.City,
		Asp:         entLoginRecord.Asp,
		IsMobile:    entLoginRecord.IsMobile,
		IsSuccess:   entLoginRecord.IsSuccess,
		ErrMessage:  entLoginRecord.ErrMessage,
		CreateAt:    entLoginRecord.CreatedAt,
	}, nil
}

func (u *UserRepoImpl) BindRole(ctx context.Context, userId int64, roleId int64) (*aggregate.User, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get Tx from context
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if !ok {
		err := cus_err.New(cus_err.InternalServerError, "get tx from context failed")
		cus_otel.Error(ctx, err.Error())
		return nil, err
	}

	// Check the role is exist
	role, err := tx.Role.Get(ctx, roleId)
	if err != nil {
		cusErr := cus_err.New(cus_err.ResourceNotFound, "role is not found", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Bind role to user
	entUser, err := tx.User.UpdateOneID(userId).SetRoles(role).Save(ctx)
	if err != nil {
		cusErr := cus_err.New(cus_err.InternalServerError, "binding role failed", err)
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to enum.UserStatus
	status, cusErr := enum.UserStatusFromInt(entUser.Status)
	if cusErr != nil {
		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to aggregate.User
	user := &aggregate.User{
		Id:                entUser.ID,
		Account:           entUser.Account,
		Password:          entUser.Password,
		PasswordFailTimes: entUser.PasswordFailTimes,
		Status:            status,
	}
	setUserLoader(u.db, user)

	return user, nil
}

func (u *UserRepoImpl) validateParameters(user *aggregate.User) *cus_err.CusError {
	if user == nil {
		return cus_err.New(cus_err.AccountPasswordError, "user is nil")
	}

	if user.Id == 0 {
		return cus_err.New(cus_err.AccountPasswordError, "user id is required")
	}

	if user.Status.Int() == 0 {
		return cus_err.New(cus_err.AccountPasswordError, "user status is required")
	}

	return nil
}

func setUserLoader(db db.Database, domainUser *aggregate.User) {
	domainUser.SetLoginRecordLoader(func(ctx context.Context) (*entity.LoginRecord, *cus_err.CusError) {
		// Start trace
		ctx, span := cus_otel.StartTrace(ctx)
		defer span.End()

		// Get client with transaction if exists.
		var client *ent.Client
		tx, ok := db.GetTx(ctx).(*ent.Tx)
		if ok {
			client = tx.Client()
		} else {
			client = db.GetConn(ctx).(*ent.Client)
		}

		// Find last login record, if not found, return error.
		entLoginRecord, err := client.User.
			Query().
			Where(user.ID(domainUser.Id)).
			QueryLoginRecords().
			Order(ent.Desc(loginrecord.FieldCreatedAt)).
			First(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				cusErr := cus_err.New(cus_err.ResourceNotFound, "login record not found", err)
				cus_otel.Error(ctx, cusErr.Error())
				return nil, cusErr
			}
			cusErr := cus_err.New(cus_err.InternalServerError, "find last login record failed", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}

		// Map to entity.LoginRecord.
		return &entity.LoginRecord{
			Id:        entLoginRecord.ID,
			Browser:   entLoginRecord.Browser,
			Ip:        entLoginRecord.IP,
			Os:        entLoginRecord.Os,
			Country:   entLoginRecord.Country,
			City:      entLoginRecord.City,
			IsSuccess: entLoginRecord.IsSuccess,
			CreateAt:  entLoginRecord.CreatedAt,
		}, nil
	})

	domainUser.SetRoleLoader(func(ctx context.Context) (*entity.Role, *cus_err.CusError) {
		// Start trace
		ctx, span := cus_otel.StartTrace(ctx)
		defer span.End()

		// Get client with transaction if exists.
		var client *ent.Client
		tx, ok := db.GetTx(ctx).(*ent.Tx)
		if ok {
			client = tx.Client()
		} else {
			client = db.GetConn(ctx).(*ent.Client)
		}

		// Find role, if not found, return error.
		entRole, err := client.User.
			Query().
			Where(user.ID(domainUser.Id)).
			QueryRoles().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				cusErr := cus_err.New(cus_err.ResourceNotFound, "role not found", err)
				cus_otel.Error(ctx, cusErr.Error())
				return nil, cusErr
			}
			cusErr := cus_err.New(cus_err.InternalServerError, "find role failed", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}

		// Map to entity.Role
		clientType, cusErr := enum.ClientTypeFromId(entRole.ClientType)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		return &entity.Role{
			Id:          entRole.ID,
			Name:        entRole.Name,
			Permissions: entRole.Permissions,
			ClientType:  clientType,
		}, nil

	})

	domainUser.SetClientLoader(func(ctx context.Context) (*aggregate.Client, *cus_err.CusError) {
		// Start trace
		ctx, span := cus_otel.StartTrace(ctx)
		defer span.End()

		// Get client with transaction if exists.
		var client *ent.Client
		tx, ok := db.GetTx(ctx).(*ent.Tx)
		if ok {
			client = tx.Client()
		} else {
			client = db.GetConn(ctx).(*ent.Client)
		}

		// Find client, if not found, return error.
		entClient, err := client.User.
			Query().
			Where(user.ID(domainUser.Id)).
			QueryAuthClients().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				cusErr := cus_err.New(cus_err.ResourceNotFound, "client not found", err)
				cus_otel.Error(ctx, cusErr.Error())
				return nil, cusErr
			}
			cusErr := cus_err.New(cus_err.InternalServerError, "find client failed", err)
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}

		// Map to aggregate.Client
		clientType, cusErr := enum.ClientTypeFromId(entClient.ClientType)
		if cusErr != nil {
			cus_otel.Error(ctx, cusErr.Error())
			return nil, cusErr
		}
		domainClient := &aggregate.Client{
			Id:               entClient.ID,
			MerchantId:       entClient.MerchantID,
			ClientType:       clientType,
			Secret:           entClient.Secret,
			Active:           entClient.Active,
			TokenExpireSecs:  entClient.TokenExpireSecs,
			LoginFailedTimes: entClient.LoginFailedTimes,
		}
		setClientLoader(db, domainClient)
		return domainClient, nil
	})

}

func (u *UserRepoImpl) CheckAccountExistence(ctx context.Context, account string) (bool, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	conn := u.db.GetConn(ctx).(*ent.Client)

	// Check the account is exist
	exist, err := conn.User.Query().Where(user.Account(account)).Exist(ctx)
	if err != nil {
		return true, cus_err.New(cus_err.InternalServerError, "failed to query error", err)
	}
	return exist, nil
}

func (u *UserRepoImpl) GetLastLoginRecord(ctx context.Context, userId int64) (*entity.LoginRecord, *cus_err.CusError) {
	// Start trace
	ctx, span := cus_otel.StartTrace(ctx)
	defer span.End()

	// Get client with transaction if exists
	var client *ent.Client
	tx, ok := u.db.GetTx(ctx).(*ent.Tx)
	if ok {
		client = tx.Client()
	} else {
		client = u.db.GetConn(ctx).(*ent.Client)
	}

	// Find last successful login record
	entLoginRecord, err := client.LoginRecord.Query().
		Where(
			loginrecord.And(
				loginrecord.HasUsersWith(user.ID(userId)),
				loginrecord.IsSuccess(true),
			),
		).
		Order(ent.Desc(loginrecord.FieldCreatedAt)).
		First(ctx)

	if err != nil {
		var cusErr *cus_err.CusError
		if ent.IsNotFound(err) {
			cusErr = cus_err.New(cus_err.ResourceNotFound, "last login record not found", err)
		} else {
			cusErr = cus_err.New(cus_err.InternalServerError, "find last login record failed", err)
		}

		cus_otel.Error(ctx, cusErr.Error())
		return nil, cusErr
	}

	// Map to entity.LoginRecord
	return &entity.LoginRecord{
		Id:          entLoginRecord.ID,
		Browser:     entLoginRecord.Browser,
		BrowserVer:  entLoginRecord.BrowserVer,
		Ip:          entLoginRecord.IP,
		Os:          entLoginRecord.Os,
		Platform:    entLoginRecord.Platform,
		Country:     entLoginRecord.Country,
		CountryCode: entLoginRecord.CountryCode,
		City:        entLoginRecord.City,
		Asp:         entLoginRecord.Asp,
		IsMobile:    entLoginRecord.IsMobile,
		IsSuccess:   entLoginRecord.IsSuccess,
		ErrMessage:  entLoginRecord.ErrMessage,
		CreateAt:    entLoginRecord.CreatedAt,
	}, nil
}
