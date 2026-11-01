package server_test

import (
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/server"
	"rcon-scheduler-panel/internal/domain/server/mocks"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupRouter(mockSVC *mocks.MockService) *gin.Engine {
	cfg := &config.Config{JWTSecret: "test"}
	h := server.NewHandler(mockSVC, cfg)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h.RegisterRoutes(r.Group("/api"))
	return r
}

func TestServerHandler_GetById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	expectedServer := &model.Server{
		ID:       1,
		Name:     "Test Server",
		Address:  "localhost",
		Port:     25575,
		Password: "test12345",
		Protocol: &model.Protocol{
			ID:   1,
			Name: "RCON",
		},
	}

	t.Run("200 - Server Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(1)).
			Return(expectedServer, nil).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers/1", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers/abc", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("404 - Server Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(999)).
			Return(nil, errs.ErrNotFound.WithMessage("server not found")).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers/999", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestServerHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	expectedServer := &model.Server{
		ID:       1,
		Name:     "Test Server",
		Address:  "localhost",
		Port:     25575,
		Password: "test12345",
		Protocol: &model.Protocol{
			ID:   1,
			Name: "RCON",
		},
	}

	t.Run("200 - Servers Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		mockSVC.EXPECT().
			GetAll(mock.Anything).
			Return([]model.Server{*expectedServer}, nil)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/servers", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}