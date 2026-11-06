package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/user"
	"rcon-scheduler-panel/internal/domain/user/mocks"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupRouter(mockSVC *mocks.MockService) *gin.Engine {
	cfg := &config.Config{JWTSecret: "test"}
	h := user.NewHandler(mockSVC, cfg)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h.RegisterRoutes(r.Group("/api"))
	return r
}

func TestUserHandler_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	expectedUser := &model.User{
		ID:        1,
		Username:  "test",
		Password:  "test1234",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("200 - User Exist (and auth is correct)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(1)).
			Return(expectedUser, nil).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool               `json:"success"`
			Data    response.UserShort `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(1), resp.Data.ID)
		assert.Equal(t, "test", resp.Data.Username)
		assert.NotContains(t, w.Body.String(), expectedUser.Password)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401 - Unauthorized", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool `json:"success"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Success)
		assert.NotContains(t, w.Body.String(), expectedUser.Password)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
