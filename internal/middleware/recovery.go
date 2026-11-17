package middleware

import (
	"net/http"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/transport/http/respond"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Recovery() gin.RecoveryFunc {
	return func(c *gin.Context, recovered any) {
		log.Error().
            Interface("panic", recovered).
            Stack().
            Msg("panic recovered")

		if c.Writer.Written() {
			c.Abort()
			return
		}

		respond.Problem(c, newProblem(c, http.StatusInternalServerError, errs.ErrInternalServer))
	}
}
