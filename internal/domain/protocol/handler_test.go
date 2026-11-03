package protocol_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/protocol"
	"rcon-scheduler-panel/internal/domain/protocol/mocks"
	"rcon-scheduler-panel/internal/dto/response"
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
	h := protocol.NewHandler(mockSVC, cfg)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h.RegisterRoutes(r.Group("/api"))
	return r
}

func TestProtocolHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	description := "RCON protocol for Source Engine's games"
	expectedProtocol := &model.Protocol{
		ID:          1,
		Name:        "Source",
		Description: &description,
	}

	t.Run("200 - Protocol Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(1)).
			Return(expectedProtocol, nil).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols/1", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool `json:"success"`
			Data    response.Protocol `json:"data"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedProtocol.ID, resp.Data.ID)
		assert.Equal(t, expectedProtocol.Name, resp.Data.Name)
		assert.Equal(t, *expectedProtocol.Description, resp.Data.Description)
	})

	t.Run("401 - Unauthorized", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool `json:"success"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Success)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("404 - Protocol not found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetByID(mock.Anything, int64(999)).
			Return(nil, errs.ErrNotFound.WithMessage("server not found")).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols/999", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProtocolHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	description := "RCON protocol for Source Engine's games"
	expectedProtocols := []model.Protocol{
		{
			ID:          1,
			Name:        "Source",
			Description: &description,
		},
		{
			ID:   2,
			Name: "Minecraft",
		},
	}

	t.Run("200 - Protocols Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetAll(mock.Anything).
			Return(expectedProtocols, nil).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool                `json:"success"`
			Data    []response.Protocol `json:"data"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, http.StatusOK, w.Code)
		require.Len(t, resp.Data, len(expectedProtocols))
		assert.Equal(t, expectedProtocols[0].ID, resp.Data[0].ID)
		assert.Equal(t, expectedProtocols[0].Name, resp.Data[0].Name)
		assert.Equal(t, *expectedProtocols[0].Description, resp.Data[0].Description)
		assert.Equal(t, expectedProtocols[1].ID, resp.Data[1].ID)
		assert.Equal(t, expectedProtocols[1].Name, resp.Data[1].Name)
		assert.Empty(t, resp.Data[1].Description)
	})

	t.Run("200 - No Protocols", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetAll(mock.Anything).
			Return([]model.Protocol{}, nil).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool                `json:"success"`
			Data    []response.Protocol `json:"data"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, resp.Data)
	})

	t.Run("401 - Unauthorized", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool `json:"success"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Success)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("500 - Service error", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			GetAll(mock.Anything).
			Return(nil, errors.New("database is down")).
			Once()

		r := setupRouter(mockSVC)

		req, _ := http.NewRequest(http.MethodGet, "/api/protocols", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp struct {
			Success bool `json:"success"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.False(t, resp.Success)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "database is down")
	})
}
