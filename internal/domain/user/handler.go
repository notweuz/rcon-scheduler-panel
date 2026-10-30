package user

import (
	"context"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/transport/http/respond"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type Handler struct {
	service Service
	cfg     *config.Config
}

func NewHandler(service Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		respond.Error(c, err)
		return
	}
	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, response.UserShort{ID: user.ID, Username: user.Username})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/users", middleware.Auth(h.cfg))
	auth.POST("/@me", h.GetMe)
}
