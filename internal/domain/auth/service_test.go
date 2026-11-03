package auth_test

import (
	"context"
	"errors"
	authUtil "rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/auth"
	"rcon-scheduler-panel/internal/domain/user"
	"rcon-scheduler-panel/internal/domain/user/mocks"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testSecret = "test"

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	creds := request.AuthCredentials{Username: "Admin", Password: "test1234"}

	setID := func(id int64) func(ctx context.Context, u *model.User) {
		return func(ctx context.Context, u *model.User) {
			u.ID = id
		}
	}

	t.Run("Registration enabled - Count is not called", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Run(setID(5)).
			Return(nil).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: true})

		token, err := svc.Register(ctx, creds)
		require.NoError(t, err)

		userID, err := authUtil.ParseToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, int64(5), userID)
	})

	t.Run("Registration disabled - first user can register", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Count(mock.Anything).
			Return(int64(0), nil).
			Once()
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Run(setID(1)).
			Return(nil).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: false})

		token, err := svc.Register(ctx, creds)
		require.NoError(t, err)

		userID, err := authUtil.ParseToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, int64(1), userID)
	})

	t.Run("Registration disabled - users exist", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Count(mock.Anything).
			Return(int64(1), nil).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: false})

		token, err := svc.Register(ctx, creds)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrForbidden)
	})

	t.Run("Registration disabled - Count failed", func(t *testing.T) {
		dbErr := errors.New("database is down")

		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Count(mock.Anything).
			Return(int64(0), dbErr).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: false})

		token, err := svc.Register(ctx, creds)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("Username is lowercased and password is hashed", func(t *testing.T) {
		var created *model.User

		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, u *model.User) {
				u.ID = 1
				created = u
			}).
			Return(nil).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: true})

		_, err := svc.Register(ctx, creds)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, "admin", created.Username)
		assert.NotEqual(t, creds.Password, created.Password)
		assert.True(t, authUtil.VerifyPasswordHash(created.Password, creds.Password))
	})

	t.Run("User already exists", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errs.ErrConflict.WithMessage("user already exists")).
			Once()

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: true})

		token, err := svc.Register(ctx, creds)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrConflict)
	})

	t.Run("Password too long for bcrypt", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		svc := auth.NewService(mockSVC, &config.Config{JWTSecret: testSecret, AllowRegistration: true})

		tooLong := request.AuthCredentials{Username: "admin", Password: strings.Repeat("я", 40)}

		token, err := svc.Register(ctx, tooLong)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrBadRequest)
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{JWTSecret: testSecret}
	creds := request.AuthCredentials{Username: "Admin", Password: "test1234"}

	hash, err := authUtil.GeneratePasswordHash(creds.Password)
	require.NoError(t, err)

	existingUser := &model.User{
		ID:       7,
		Username: "admin",
		Password: hash,
	}

	t.Run("Correct credentials", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByUsername(mock.Anything, "admin").
			Return(existingUser, nil).
			Once()

		svc := auth.NewService(mockSVC, cfg)

		token, err := svc.Login(ctx, creds)
		require.NoError(t, err)

		userID, err := authUtil.ParseToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, existingUser.ID, userID)
	})

	t.Run("Wrong password", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByUsername(mock.Anything, "admin").
			Return(existingUser, nil).
			Once()

		svc := auth.NewService(mockSVC, cfg)

		token, err := svc.Login(ctx, request.AuthCredentials{Username: "admin", Password: "wrongpass"})
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)
	})

	t.Run("User not found - 401 (no token provided)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		mockRepo.EXPECT().
			GetByUsername(mock.Anything, "ghost").
			Return(nil, user.ErrNotFound).
			Once()

		svc := auth.NewService(user.NewService(mockRepo), cfg)

		token, err := svc.Login(ctx, request.AuthCredentials{Username: "ghost", Password: "test1234"})
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrUnauthorized)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, errs.Unauthenticated, appErr.Kind)
		assert.Equal(t, "invalid credentials", appErr.Message)
	})

	t.Run("Service error is not masked as 401", func(t *testing.T) {
		dbErr := errors.New("database is down")

		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByUsername(mock.Anything, "admin").
			Return(nil, dbErr).
			Once()

		svc := auth.NewService(mockSVC, cfg)

		token, err := svc.Login(ctx, creds)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, dbErr)
		assert.NotErrorIs(t, err, errs.ErrUnauthorized)
	})
}
