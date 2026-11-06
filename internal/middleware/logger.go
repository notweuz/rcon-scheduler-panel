package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		var ev *zerolog.Event

		switch {
		case status >= 500:
			ev = log.Error()
		case status >= 400:
			ev = log.Warn()
		default:
			ev = log.Info()
		}

		ev.Str("ip", c.ClientIP()).Str("method", c.Request.Method).Str("path", c.Request.URL.Path).Int("status", c.Writer.Status()).Dur("latency", time.Since(start)).Msg("http request")
	}
}
