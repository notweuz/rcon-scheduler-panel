package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/errs"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTSecret: "test"}

	validToken, err := auth.GenerateToken(7, cfg.JWTSecret)
	require.NoError(t, err)

	foreignToken, err := auth.GenerateToken(7, "other-secret")
	require.NoError(t, err)

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantMsg    string
	}{
		{"valid token", "Bearer " + validToken, http.StatusOK, ""},
		{"lowercase scheme", "bearer " + validToken, http.StatusOK, ""},
		{"missing header", "", http.StatusUnauthorized, ErrTokenMissing.Message},
		{"basic scheme", "Basic " + validToken, http.StatusUnauthorized, "invalid authorization header"},
		{"no scheme", validToken, http.StatusUnauthorized, "invalid authorization header"},
		{"bearer without token", "Bearer ", http.StatusUnauthorized, "invalid authorization header"},
		{"bearer without space", "Bearer", http.StatusUnauthorized, "invalid authorization header"},
		{"wrong secret", "Bearer " + foreignToken, http.StatusUnauthorized, auth.ErrJWTSignatureFailed.Message},
		{"garbage token", "Bearer not-a-jwt", http.StatusUnauthorized, errs.ErrUnauthorized.Message},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false

			r := gin.New()
			r.Use(ErrorHandler())
			r.GET("/", Auth(cfg), func(c *gin.Context) {
				handlerCalled = true
				userID, err := GetUserIDFromContext(c)
				require.NoError(t, err)
				c.JSON(http.StatusOK, gin.H{"user_id": userID})
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				assert.True(t, handlerCalled)

				var resp struct {
					UserID int64 `json:"user_id"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, int64(7), resp.UserID)
				return
			}

			assert.False(t, handlerCalled)

			var resp struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantMsg, resp.Error.Message)
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		value   any
		set     bool
		want    int64
		wantErr bool
	}{
		{"valid", int64(7), true, 7, false},
		{"missing", nil, false, 0, true},
		{"wrong type", "7", true, 0, true},
		{"int instead of int64", 7, true, 0, true},
		{"zero", int64(0), true, 0, true},
		{"negative", int64(-1), true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.set {
				c.Set(userIDContextKey, tt.value)
			}

			userID, err := GetUserIDFromContext(c)
			assert.Equal(t, tt.want, userID)
			if tt.wantErr {
				assert.ErrorIs(t, err, errs.ErrInternalServer)
				return
			}
			assert.NoError(t, err)
		})
	}
}
