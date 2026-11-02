package server

import (
	"context"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/dto/path"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/transport/http/respond"
	"rcon-scheduler-panel/internal/validation"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Create(ctx context.Context, srv *model.Server) error
	GetByID(ctx context.Context, id int64) (*model.Server, error)
	GetAll(ctx context.Context) ([]model.Server, error)
	Delete(ctx context.Context, id int64) error
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

func (h *Handler) Create(c *gin.Context) {
	var serverReq request.CreateServer
	if err := c.ShouldBindJSON(&serverReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	srv := &model.Server{
		Name:        serverReq.Name,
		Address:     serverReq.Address,
		Port:        serverReq.Port,
		Password:    serverReq.Password,
		ProtocolID:  serverReq.ProtocolID,
	}

	if err := h.service.Create(c.Request.Context(), srv); err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, response.ServerShortFromModel(*srv))
}

func (h *Handler) GetByID(c *gin.Context) {
	var serverID path.ServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	server, err := h.service.GetByID(c.Request.Context(), serverID.ID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.ServerShortFromModel(*server))
}

func (h *Handler) GetAll(c *gin.Context) {
	servers, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	serverDTOs := make([]response.ServerShort, len(servers))
	for i, srv := range servers {
		serverDTOs[i] = response.ServerShortFromModel(srv)
	}
	respond.OK(c, serverDTOs)
}

func (h *Handler) Delete(c *gin.Context) {
	var serverID path.ServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	err := h.service.Delete(c.Request.Context(), serverID.ID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	servers := r.Group("/servers", middleware.Auth(h.cfg))
	{
		servers.GET("/:id", h.GetByID)
		servers.DELETE("/:id", h.Delete)
		servers.GET("", h.GetAll)
		servers.POST("", h.Create)
	}
}