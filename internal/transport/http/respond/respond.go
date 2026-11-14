package respond

import (
	"net/http"
	"rcon-scheduler-panel/internal/dto/response"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

func Problem(c *gin.Context, p response.Problem) {
	c.Header("Content-Type", response.ProblemContentType)
	c.AbortWithStatusJSON(p.Status, p)
}
