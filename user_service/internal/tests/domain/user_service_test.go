package domain_test

import (
	"context"
	"go_micro_service_api/pkg/db"
	"go_micro_service_api/pkg/enum"
	"go_micro_service_api/user_service/internal/domain/aggregate"
	"go_micro_service_api/user_service/internal/domain/entity"
	"go_micro_service_api/user_service/internal/domain/repository"
	"go_micro_service_api/user_service/internal/domain/service"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent"
	"go_micro_service_api/user_service/internal/infrastructure/ent_impl/ent/profile"
	"go_micro_service_api/user_service/internal/tests"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserService() (userService *service.UserService, userRepo repository.UserRepo, db db.Database, closeFunc func()) {
	db = tests.NewMemoryDB()
	userRepo = ent_impl.NewUserRepo(db)

	return service.NewUserService(userRepo), userRepo, db, closeFunc
}

func TestCreateProfile(t *testing.T) {
	service, _, db, _ := setupUserService()

	tcs := []struct {
		name    string
		profile entity.Profile
		fields  []int
		values  []string
		wantErr bool
	}{
		{
			name: "Create email",
			profile: entity.Profile{
				Email: "test@gmail.com",
			},
			fields:  []int{enum.ProfileKey.Email.ID},
			values:  []string{"test@gmail.com"},
			wantErr: false,
		},
		{
			name: "Create phone",
			profile: entity.Profile{
				CountryCode:  "081",
				MobileNumber: "123456789",
			},
			fields: []int{
				enum.ProfileKey.CountryCode.ID,
				enum.ProfileKey.MobileNumber.ID,
			},
			values: []string{
				"081",
				"123456789",
			},
			wantErr: false,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Assign
			ctx := context.Background()
			ctx, err := db.Begin(ctx)
			require.Nil(t, err)

			client := db.GetClient(ctx).(*ent.Client)
			u := &aggregate.User{
				Profile: tc.profile,
			}

			// Act
			_, err = service.CreateProfile(ctx, u)

			// Assert
			assert.Equal(t, tc.wantErr, err != nil)
			if !tc.wantErr {
				for i, field := range tc.fields {
					exist, err := client.Profile.
						Query().
						Where(
							profile.UserIDEQ(int(u.ID)),
							profile.KeyEQ(field),
							profile.ValueEQ(tc.values[i]),
						).
						Exist(ctx)

					require.Nil(t, err)
					assert.True(t, exist)
				}
			}

			_, err = db.Rollback(ctx)
			require.Nil(t, err)
		})
	}
}

func TestGetProfile(t *testing.T) {
	service, repo, db, _ := setupUserService()
	ctx := context.Background()
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	_, err = repo.CreateProfile(ctx, &aggregate.User{
		Profile: entity.Profile{
			Email:        "test@gmail.com",
			CountryCode:  "081",
			MobileNumber: "123456789",
		},
	})
	require.Nil(t, err)
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	tcs := []struct {
		name    string
		profile entity.Profile
		fields  []int
		values  []string
		wantErr error
	}{
		{
			name: "Get email",
			profile: entity.Profile{
				Email: "test@gmail.com",
			},
			fields:  []int{enum.ProfileKey.Email.ID},
			values:  []string{"test@gmail.com"},
			wantErr: nil,
		},
		{
			name: "Get phone",
			profile: entity.Profile{
				CountryCode:  "081",
				MobileNumber: "123456789",
			},
			fields: []int{
				enum.ProfileKey.CountryCode.ID,
				enum.ProfileKey.MobileNumber.ID,
			},
			values: []string{
				"081",
				"123456789",
			},
			wantErr: nil,
		},
		// TODO: fail case
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			u := &aggregate.User{
				Profile: entity.Profile{},
			}

			// Act
			u, err = service.GetProfile(ctx, u, tc.fields)

			// Assert
			assert.Equal(t, tc.wantErr != nil, err != nil)
			if tc.wantErr != nil {
				for _, field := range tc.fields {
					switch field {
					case enum.ProfileKey.Email.ID:
						assert.Equal(t, tc.profile.Email, u.Profile.Email)
					case enum.ProfileKey.CountryCode.ID:
						assert.Equal(t, tc.profile.CountryCode, u.Profile.CountryCode)
					case enum.ProfileKey.MobileNumber.ID:
						assert.Equal(t, tc.profile.MobileNumber, u.Profile.MobileNumber)
					}
				}
			}
		})
	}
}

func TestCheckMobileExistence(t *testing.T) {
	service, repo, db, _ := setupUserService()
	// Assign
	ctx := context.Background()
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	_, err = repo.CreateProfile(ctx, &aggregate.User{
		Profile: entity.Profile{
			CountryCode:  "081",
			MobileNumber: "123456789",
		},
	})
	require.Nil(t, err)
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	type Q struct {
		CountryCode  string
		MobileNumber string
	}

	tcs := []struct {
		name    string
		query   Q
		want    bool
		wantErr error
	}{
		{
			name: "Exist",
			query: Q{
				CountryCode:  "081",
				MobileNumber: "123456789",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "NonExist",
			query: Q{
				CountryCode:  "081",
				MobileNumber: "1289",
			},
			want:    false,
			wantErr: nil,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			exist, err := service.CheckMobileExistence(ctx, tc.query.CountryCode, tc.query.MobileNumber)

			// Assert
			assert.Equal(t, tc.wantErr == nil, err == nil)
			assert.Equal(t, tc.want, exist)
		})
	}
}

func TestCheckEmailExistence(t *testing.T) {
	service, repo, db, _ := setupUserService()
	// Assign
	ctx := context.Background()
	ctx, err := db.Begin(ctx)
	require.Nil(t, err)
	_, err = repo.CreateProfile(ctx, &aggregate.User{
		Profile: entity.Profile{
			Email: "test@gmail.com",
		},
	})
	require.Nil(t, err)
	ctx, err = db.Commit(ctx)
	require.Nil(t, err)

	type Q struct {
		Email string
	}

	tcs := []struct {
		name    string
		query   Q
		want    bool
		wantErr error
	}{
		{
			name: "Exist",
			query: Q{
				Email: "test@gmail.com",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "NonExist",
			query: Q{
				Email: "test2@gmail.com",
			},
			want:    false,
			wantErr: nil,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			exist, err := service.CheckEmailExistence(ctx, tc.query.Email)

			// Assert
			assert.Equal(t, tc.wantErr == nil, err == nil)
			assert.Equal(t, tc.want, exist)
		})
	}
}
