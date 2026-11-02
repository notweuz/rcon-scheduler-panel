package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/errs"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorHandlerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want int
	}{
		{"bad request", errs.ErrBadRequest, http.StatusBadRequest},
		{"method not allowed", errs.ErrMethodNotAllowed, http.StatusMethodNotAllowed},
		{"wrapped not found", fmt.Errorf("ctx: %w", errs.ErrNotFound), http.StatusNotFound},
		{"conflict", errs.ErrConflict, http.StatusConflict},
		{"bad gateway", errs.ErrBadGateway, http.StatusBadGateway},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandler())
			r.GET("/", func(c *gin.Context) { _ = c.Error(tt.err) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

			if w.Code != tt.want {
				t.Errorf("got status %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestErrorHandlerMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"internal is hidden", errs.ErrInternalServer.WithMessage("db password is wrong"), errs.ErrInternalServer.Message},
		{"unknown is hidden", errors.New("boom"), errs.ErrInternalServer.Message},
		{"bad gateway is shown", errs.ErrBadGateway.WithMessage("game server is down"), "game server is down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandler())
			r.GET("/", func(c *gin.Context) { _ = c.Error(tt.err) })

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

			var resp struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp.Error.Message != tt.want {
				t.Errorf("got message %q, want %q", resp.Error.Message, tt.want)
			}
		})
	}
}
