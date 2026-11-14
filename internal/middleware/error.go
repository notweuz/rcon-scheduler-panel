package middleware

import (
	"errors"
	"net/http"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/transport/http/respond"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var statusByKind = map[errs.Kind]int{
	errs.BadRequest:       http.StatusBadRequest,
	errs.Unauthenticated:  http.StatusUnauthorized,
	errs.Forbidden:        http.StatusForbidden,
	errs.NotFound:         http.StatusNotFound,
	errs.MethodNotAllowed: http.StatusMethodNotAllowed,
	errs.Conflict:         http.StatusConflict,
	errs.Internal:         http.StatusInternalServerError,
	errs.BadGateway:       http.StatusBadGateway,
}

func httpStatus(k errs.Kind) int {
	if s, ok := statusByKind[k]; ok {
		return s
	}
	return http.StatusInternalServerError
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		var appErr *errs.AppError
		if !errors.As(err, &appErr) {
			appErr = errs.ErrInternalServer
		}

		status := httpStatus(appErr.Kind)
		if status >= http.StatusInternalServerError {
			log.Error().
				Err(err).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Msg("request failed")
		}

		if status == http.StatusInternalServerError {
			appErr = errs.ErrInternalServer
		}

		if status == http.StatusUnauthorized {
			c.Header("WWW-Authenticate", `Bearer realm="api"`)
		}

		respond.Problem(c, newProblem(c, status, appErr))
	}
}

func newProblem(c *gin.Context, status int, e *errs.AppError) response.Problem {
	p := response.Problem{
		Title:    http.StatusText(status),
		Status:   status,
		Detail:   e.Message,
		Instance: c.Request.URL.Path,
		Code:     e.Code,
	}
	if len(e.Violations) > 0 {
		p.Errors = make([]response.InvalidParam, len(e.Violations))
		for i, v := range e.Violations {
			p.Errors[i] = response.InvalidParam{Field: v.Field, Code: v.Code, Param: v.Param}
		}
	}
	return p
}
