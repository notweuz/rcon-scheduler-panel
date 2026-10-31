package respond

import (
	"net/http"
	"rcon-scheduler-panel/internal/dto/response"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, response.Response{
		Success: true,
		Data:    data,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}
