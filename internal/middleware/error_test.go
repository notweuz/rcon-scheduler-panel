package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/errs"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			assert.Equal(t, tt.want, w.Code)
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
				Detail string `json:"detail"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.want, resp.Detail)
		})
	}
}

func TestErrorHandlerProblemDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/servers/:id", func(c *gin.Context) {
		_ = c.Error(errs.ErrNotFound.WithCode("SERVER_NOT_FOUND").WithMessage("server not found"))
	})
	r.POST("/validate", func(c *gin.Context) {
		_ = c.Error(errs.ErrBadRequest.WithCode("VALIDATION_FAILED").WithViolations([]errs.Violation{
			{Field: "steps[0].command", Code: "required"},
			{Field: "name", Code: "min", Param: "3"},
		}))
	})
	r.GET("/private", func(c *gin.Context) { _ = c.Error(errs.ErrUnauthorized) })

	t.Run("problem body and content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/servers/42?x=1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "application/problem+json", w.Header().Get("Content-Type"))

		var p response.Problem
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		assert.Equal(t, response.Problem{
			Title:    "Not Found",
			Status:   http.StatusNotFound,
			Detail:   "server not found",
			Instance: "/servers/42",
			Code:     "SERVER_NOT_FOUND",
		}, p)
	})

	t.Run("validation errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/validate", nil))

		var p response.Problem
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		assert.Equal(t, []response.InvalidParam{
			{Field: "steps[0].command", Code: "required"},
			{Field: "name", Code: "min", Param: "3"},
		}, p.Errors)
	})

	t.Run("401 sets WWW-Authenticate", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/private", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, `Bearer realm="api"`, w.Header().Get("WWW-Authenticate"))
	})

	t.Run("internal error hides code and detail", func(t *testing.T) {
		rr := gin.New()
		rr.Use(ErrorHandler())
		rr.GET("/", func(c *gin.Context) {
			_ = c.Error(errs.ErrInternalServer.WithCode("DB_DOWN").WithMessage("secret"))
		})
		w := httptest.NewRecorder()
		rr.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		var p response.Problem
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		assert.Equal(t, errs.ErrInternalServer.Code, p.Code)
		assert.Equal(t, errs.ErrInternalServer.Message, p.Detail)
	})
}
