package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/domain/auth"
	"rcon-scheduler-panel/internal/domain/auth/mocks"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/validation"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupRouter(mockSVC *mocks.MockService) *gin.Engine {
	h := auth.NewHandler(mockSVC)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h.RegisterRoutes(r.Group("/api"))
	return r
}

func newAuthRequest(path string, body []byte) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func mustJSON(t *testing.T, v any) []byte {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	const path = "/api/auth/login"
	validReq := request.AuthCredentials{Username: "admin", Password: "test1234"}

	t.Run("200 - Logged In", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Login(mock.Anything, validReq).
			Return("jwt-token", nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                  `json:"success"`
			Data    response.AuthResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "jwt-token", resp.Data.Token)
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, []byte(`{"username":`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (empty body)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := request.AuthCredentials{Username: "ab", Password: "short"}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, invalid)))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Fields, "username")
		assert.Contains(t, resp.Error.Fields, "password")
	})

	t.Run("401 - Invalid Credentials", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Login(mock.Anything, validReq).
			Return("", errs.ErrUnauthorized.WithMessage("invalid credentials")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		require.Equal(t, http.StatusUnauthorized, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "invalid credentials", resp.Error.Message)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Login(mock.Anything, validReq).
			Return("", errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	const path = "/api/auth/register"
	validReq := request.AuthCredentials{Username: "admin", Password: "test1234"}

	t.Run("201 - Registered", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Register(mock.Anything, validReq).
			Return("jwt-token", nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		require.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool                  `json:"success"`
			Data    response.AuthResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "jwt-token", resp.Data.Token)
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, []byte(`{"username":`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := request.AuthCredentials{
			Username: "this-username-is-way-too-long",
			Password: "this-password-is-way-too-long-for-the-limit-of-40",
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, invalid)))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		require.Contains(t, resp.Error.Fields, "username")
		require.Contains(t, resp.Error.Fields, "password")
		assert.Equal(t, "max", resp.Error.Fields["username"].Code)
		assert.Equal(t, "max", resp.Error.Fields["password"].Code)
	})

	t.Run("400 - Bad Request (invalid field type)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, []byte(`{"username":123,"password":"test1234"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("403 - Forbidden (registration disabled)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Register(mock.Anything, validReq).
			Return("", errs.ErrForbidden.WithMessage("registration is disabled")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		require.Equal(t, http.StatusForbidden, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Equal(t, "registration is disabled", resp.Error.Message)
	})

	t.Run("409 - Conflict (user already exists)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Register(mock.Anything, validReq).
			Return("", errs.ErrConflict.WithMessage("user already exists")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Register(mock.Anything, validReq).
			Return("", errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newAuthRequest(path, mustJSON(t, validReq)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}
