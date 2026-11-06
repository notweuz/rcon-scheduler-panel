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
//	@Success	200	{object}	response.Response{data=response.Protocol}
//	@Failure	401	{object}	response.Response{error=response.ErrorInfo}	"Missing or invalid token"
//	@Failure	404	{object}	response.Response{error=response.ErrorInfo}	"Protocol not found"
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
//	@Success	200	{object}	response.Response{data=[]response.Protocol}
//	@Failure	401	{object}	response.Response{error=response.ErrorInfo}	"Missing or invalid token"
//	@Router		/protocols [get]
func (h *Handler) GetAll(c *gin.Context) {
	protocols, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	protocolDTOs := make([]response.Protocol, len(protocols))
	for i, p := range protocols {
		protocolDTOs[i] = response.ProtocolFromModel(p)
	}
	respond.OK(c, protocolDTOs)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	protocols := r.Group("/protocols", middleware.Auth(h.cfg))
	{
		protocols.GET("/:id", h.GetByID)
		protocols.GET("", h.GetAll)
	}
}
