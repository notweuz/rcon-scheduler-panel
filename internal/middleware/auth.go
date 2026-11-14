package middleware

import (
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/dto/header"
	"rcon-scheduler-panel/internal/errs"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ErrTokenMissing      = errs.ErrUnauthorized.WithCode("TOKEN_MISSING").WithMessage("authorization header is missing")
	ErrInvalidAuthHeader = errs.ErrUnauthorized.WithCode("INVALID_AUTHORIZATION_HEADER").WithMessage("authorization header must use the Bearer scheme")
)

const userIDContextKey = "user_id"

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var header header.Authorization
		if err := c.ShouldBindHeader(&header); err != nil {
			_ = c.Error(ErrTokenMissing.WithErr(err))
			c.Abort()
			return
		}

		scheme, token, ok := strings.Cut(header.Token, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			_ = c.Error(ErrInvalidAuthHeader)
			c.Abort()
			return
		}

		userID, err := auth.ParseToken(token, cfg.JWTSecret)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(userIDContextKey, userID)
	}
}

func GetUserIDFromContext(c *gin.Context) (int64, error) {
	value, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, errs.ErrInternalServer.WithMessage("user id is missing in context")
	}

	userID, ok := value.(int64)
	if !ok || userID <= 0 {
		return 0, errs.ErrInternalServer.WithMessage("invalid user ID in context")
	}

	return userID, nil
}
