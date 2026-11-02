package protocol

import (
	"context"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/dto/path"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/transport/http/respond"
	"rcon-scheduler-panel/internal/validation"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Create(ctx context.Context, protocol *model.Protocol) error
	GetByID(ctx context.Context, id int64) (*model.Protocol, error)
	GetAll(ctx context.Context) ([]model.Protocol, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct {
	service Service
	cfg *config.Config
}

func NewHandler(service Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var protocolReq request.CreateProtocol
	if err := c.ShouldBindJSON(&protocolReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	protocol := &model.Protocol{
		Name:        protocolReq.Name,
		Description: protocolReq.Description,
	}

	if err := h.service.Create(c.Request.Context(), protocol); err != nil {
		respond.Error(c, err)
		return
	}

	// this doesn't need dto for now because protocol doesn't content any sensitive information
	respond.Created(c, protocol)
}

func (h *Handler) GetByID(c *gin.Context) {
	var protocolID path.ProtocolID
	if err := c.ShouldBindUri(&protocolID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	protocol, err := h.service.GetByID(c.Request.Context(), protocolID.ID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, protocol)
}

func (h *Handler) GetAll(c *gin.Context) {
	protocols, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, protocols)
}

func (h *Handler) Delete(c *gin.Context) {
	var protocolID path.ProtocolID
	if err := c.ShouldBindUri(&protocolID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	err := h.service.Delete(c.Request.Context(), protocolID.ID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	protocols := r.Group("/protocols", middleware.Auth(h.cfg))
	{
		protocols.GET("/:id", h.GetByID)
		protocols.DELETE("/:id", h.Delete)
		protocols.GET("", h.GetAll)
		protocols.POST("", h.Create)
	}
}