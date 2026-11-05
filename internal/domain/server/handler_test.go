package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/server"
	"rcon-scheduler-panel/internal/domain/server/mocks"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/validation"
	"strings"
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

		var resp struct {
			Success bool                   `json:"success"`
			Data    []response.ServerShort `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Data, 1)
		assert.Equal(t, expectedServer.ID, resp.Data[0].ID)
		assert.Equal(t, expectedServer.Protocol.Name, resp.Data[0].Protocol)
		assert.NotContains(t, w.Body.String(), expectedServer.Password)
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

func TestServerHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	validReq := request.CreateServer{
		Name:       "Test Server",
		Address:    "localhost",
		Port:       25575,
		Password:   "test12345",
		ProtocolID: 1,
	}

	newRequest := func(body []byte, withToken bool) *http.Request {
		req, _ := http.NewRequest(http.MethodPost, "/api/servers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if withToken {
			req.Header.Set("Authorization", "Bearer "+testToken)
		}
		return req
	}

	mustJSON := func(v any) []byte {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return b
	}

	t.Run("201 - Server Created", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(srv *model.Server) bool {
				return srv.Name == validReq.Name &&
					srv.Address == validReq.Address &&
					srv.Port == validReq.Port &&
					srv.Password == validReq.Password &&
					srv.ProtocolID == validReq.ProtocolID
			})).
			Run(func(_ context.Context, srv *model.Server) {
				srv.ID = 42
			}).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(mustJSON(validReq), true))

		require.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool                 `json:"success"`
			Data    response.ServerShort `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(42), resp.Data.ID)
		assert.Equal(t, validReq.Name, resp.Data.Name)
		assert.Equal(t, validReq.Address, resp.Data.Address)
		assert.Equal(t, validReq.Port, resp.Data.Port)
		assert.NotContains(t, w.Body.String(), validReq.Password)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(mustJSON(validReq), false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest([]byte(`{"name":`), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := validReq
		invalid.Name = "ab"
		invalid.Password = "short"

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(mustJSON(invalid), true))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Fields, "name")
		assert.Contains(t, resp.Error.Fields, "password")
	})

	t.Run("400 - Bad Request (invalid field type)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		body := []byte(`{"name":"Test Server","address":"localhost","port":"abc","password":"test12345","protocol_id":1}`)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(body, true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("409 - Conflict (server already exists)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errs.ErrConflict.WithMessage("server already exist")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(mustJSON(validReq), true))

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest(mustJSON(validReq), true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestServerHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	validReq := request.UpdateServer{
		Name:       "Updated Server",
		Address:    "127.0.0.1",
		Port:       27015,
		Password:   "newpass123",
		ProtocolID: 2,
	}

	newRequest := func(url string, body []byte, withToken bool) *http.Request {
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if withToken {
			req.Header.Set("Authorization", "Bearer "+testToken)
		}
		return req
	}

	mustJSON := func(v any) []byte {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return b
	}

	t.Run("200 - Server Updated", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.MatchedBy(func(srv *model.Server) bool {
				return srv.ID == 1 &&
					srv.Name == validReq.Name &&
					srv.Address == validReq.Address &&
					srv.Port == validReq.Port &&
					srv.Password == validReq.Password &&
					srv.ProtocolID == validReq.ProtocolID
			})).
			Run(func(_ context.Context, srv *model.Server) {
				srv.Protocol = &model.Protocol{ID: 2, Name: "Source"}
			}).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(validReq), true))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                 `json:"success"`
			Data    response.ServerShort `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, int64(1), resp.Data.ID)
		assert.Equal(t, validReq.Name, resp.Data.Name)
		assert.Equal(t, validReq.Address, resp.Data.Address)
		assert.Equal(t, validReq.Port, resp.Data.Port)
		assert.Equal(t, "Source", resp.Data.Protocol)
		assert.NotContains(t, w.Body.String(), validReq.Password)
	})

	t.Run("200 - Password is optional", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.MatchedBy(func(srv *model.Server) bool {
				return srv.ID == 1 && srv.Password == ""
			})).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		noPassword := validReq
		noPassword.Password = ""

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(noPassword), true))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(validReq), false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/abc", mustJSON(validReq), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", []byte(`{"name":`), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		invalid := validReq
		invalid.Name = "ab"
		invalid.Password = "short"
		invalid.ProtocolID = 0

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(invalid), true))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Fields, "name")
		assert.Contains(t, resp.Error.Fields, "password")
		assert.Contains(t, resp.Error.Fields, "protocol_id")
	})

	t.Run("404 - Server Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.Anything).
			Return(errs.ErrNotFound.WithMessage("server not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/999", mustJSON(validReq), true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("409 - Conflict (server already registered)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.Anything).
			Return(errs.ErrConflict.WithMessage("server already registered")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(validReq), true))

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Update(mock.Anything, mock.Anything).
			Return(errors.New("db is down")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", mustJSON(validReq), true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "db is down")
	})
}

func TestServerHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	newRequest := func(url string, withToken bool) *http.Request {
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		if withToken {
			req.Header.Set("Authorization", "Bearer "+testToken)
		}
		return req
	}

	t.Run("204 - Server Deleted", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Delete(mock.Anything, int64(1)).
			Return(nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", true))

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/abc", true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("404 - Server Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Delete(mock.Anything, int64(999)).
			Return(errs.ErrNotFound.WithMessage("server not found")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/999", true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Delete(mock.Anything, int64(1)).
			Return(errors.New("unschedule jobs of server 1: boom")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1", true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "boom")
	})
}

func TestServerHandler_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validation.Register()

	testToken, err := auth.GenerateToken(1, "test")
	require.NoError(t, err)

	newRequest := func(url string, body []byte, withToken bool) *http.Request {
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if withToken {
			req.Header.Set("Authorization", "Bearer "+testToken)
		}
		return req
	}

	mustJSON := func(v any) []byte {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return b
	}

	validReq := request.ExecuteCommand{Command: "status"}

	t.Run("200 - Command Executed", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Execute(mock.Anything, int64(1), "status").
			Return("players : 3 humans, 0 bots", nil).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1/execute", mustJSON(validReq), true))

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool                   `json:"success"`
			Data    response.ExecuteResult `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, "players : 3 humans, 0 bots", resp.Data.Response)
	})

	t.Run("401 - Unauthorized (no token provided)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1/execute", mustJSON(validReq), false))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 - Bad Request (invalid ID in URL)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/abc/execute", mustJSON(validReq), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (malformed JSON)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1/execute", []byte(`{"command":`), true))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 - Bad Request (validation failed)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)

		r := setupRouter(mockSVC)

		tests := []struct {
			name    string
			command string
			wantTag string
		}{
			{"empty command", "", "required"},
			{"command too long", strings.Repeat("a", 1001), "max"},
		}

		for _, tt := range tests {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, newRequest("/api/servers/1/execute", mustJSON(request.ExecuteCommand{Command: tt.command}), true))

			require.Equal(t, http.StatusBadRequest, w.Code, tt.name)

			var resp response.Response
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.NotNil(t, resp.Error, tt.name)
			require.Contains(t, resp.Error.Fields, "command", tt.name)
			assert.Equal(t, tt.wantTag, resp.Error.Fields["command"].Code, tt.name)
		}
	})

	t.Run("404 - Server Not Found", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Execute(mock.Anything, int64(999), "status").
			Return("", fmt.Errorf("execute command for 999 server: %w", errs.ErrNotFound.WithMessage("server not found"))).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/999/execute", mustJSON(validReq), true))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("502 - Bad Gateway (game server is unavailable)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Execute(mock.Anything, int64(1), "status").
			Return("", errs.ErrBadGateway.WithMessage("game server is unavailable or rejected the command")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1/execute", mustJSON(validReq), true))

		require.Equal(t, http.StatusBadGateway, w.Code)

		var resp response.Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.NotNil(t, resp.Error)
		assert.Equal(t, "game server is unavailable or rejected the command", resp.Error.Message)
	})

	t.Run("500 - Internal Server Error (unexpected service error)", func(t *testing.T) {
		mockSVC := mocks.NewMockService(t)
		mockSVC.EXPECT().
			Execute(mock.Anything, int64(1), "status").
			Return("", errors.New("server is not registered in scheduler")).
			Once()

		r := setupRouter(mockSVC)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, newRequest("/api/servers/1/execute", mustJSON(validReq), true))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotContains(t, w.Body.String(), "not registered")
	})
}
