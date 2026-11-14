package protocol

import (
	"context"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/dto/path"
	"rcon-scheduler-panel/internal/dto/response"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/model"
	"rcon-scheduler-panel/internal/transport/http/respond"
	"rcon-scheduler-panel/internal/validation"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetByID(ctx context.Context, id int64) (*model.Protocol, error)
	GetAll(ctx context.Context) ([]model.Protocol, error)
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

// GetByID godoc
//
//	@Summary	Get protocol by ID
//	@Tags		protocols
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"Protocol ID"
//	@Success	200	{object}	response.Protocol
//	@Failure	401	{object}	response.Problem	"Missing or invalid token"
//	@Failure	404	{object}	response.Problem	"Protocol not found"
//	@Failure	500	{object}	response.Problem	"Internal server error"
//	@Router		/protocols/{id} [get]
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

	respond.OK(c, response.ProtocolFromModel(*protocol))
}

// GetAll godoc
//
//	@Summary	List protocols
//	@Tags		protocols
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.ProtocolList
//	@Failure	401	{object}	response.Problem	"Missing or invalid token"
//	@Failure	500	{object}	response.Problem	"Internal server error"
//	@Router		/protocols [get]
func (h *Handler) GetAll(c *gin.Context) {
	protocols, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.MapList(protocols, response.ProtocolFromModel))
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	protocols := r.Group("/protocols", middleware.Auth(h.cfg))
	{
		protocols.GET("/:id", h.GetByID)
		protocols.GET("", h.GetAll)
	}
}
