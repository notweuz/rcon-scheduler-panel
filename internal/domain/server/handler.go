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
	Update(ctx context.Context, srv *model.Server) error
	Execute(ctx context.Context, id int64, command string) (string, error)
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

// Create godoc
//
//	@Summary		Create server
//	@Description	Adds an RCON server to the panel.
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			server	body		request.CreateServer	true	"Server data"
//	@Success		201		{object}	response.ServerShort
//	@Failure		400		{object}	response.Problem	"Validation error"
//	@Failure		401		{object}	response.Problem	"Missing or invalid token"
//	@Failure		409		{object}	response.Problem	"Server already exists"
//	@Failure		500		{object}	response.Problem	"Internal server error"
//	@Router			/servers [post]
func (h *Handler) Create(c *gin.Context) {
	var serverReq request.CreateServer
	if err := c.ShouldBindJSON(&serverReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	srv := &model.Server{
		Name:       serverReq.Name,
		Address:    serverReq.Address,
		Port:       serverReq.Port,
		Password:   serverReq.Password,
		ProtocolID: serverReq.ProtocolID,
	}

	if err := h.service.Create(c.Request.Context(), srv); err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, response.ServerShortFromModel(*srv))
}

// GetByID godoc
//
//	@Summary	Get server by ID
//	@Tags		servers
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		int	true	"Server ID"
//	@Success	200	{object}	response.ServerShort
//	@Failure	401	{object}	response.Problem	"Missing or invalid token"
//	@Failure	404	{object}	response.Problem	"Server not found"
//	@Failure	500	{object}	response.Problem	"Internal server error"
//	@Router		/servers/{id} [get]
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

// GetAll godoc
//
//	@Summary	List servers
//	@Tags		servers
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	response.ServerList
//	@Failure	401	{object}	response.Problem	"Missing or invalid token"
//	@Failure	500	{object}	response.Problem	"Internal server error"
//	@Router		/servers [get]
func (h *Handler) GetAll(c *gin.Context) {
	servers, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.MapList(servers, response.ServerShortFromModel))
}

// Update godoc
//
//	@Summary		Update server
//	@Description	Replaces server data. Empty password keeps the current one.
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Server ID"
//	@Param			server	body		request.UpdateServer	true	"Server data"
//	@Success		200		{object}	response.ServerShort
//	@Failure		400		{object}	response.Problem	"Validation error"
//	@Failure		401		{object}	response.Problem	"Missing or invalid token"
//	@Failure		404		{object}	response.Problem	"Server or protocol not found"
//	@Failure		409		{object}	response.Problem	"Server already registered"
//	@Failure		500		{object}	response.Problem	"Internal server error"
//	@Router			/servers/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var serverID path.ServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	var serverReq request.UpdateServer
	if err := c.ShouldBindJSON(&serverReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	srv := &model.Server{
		ID:         serverID.ID,
		Name:       serverReq.Name,
		Address:    serverReq.Address,
		Port:       serverReq.Port,
		Password:   serverReq.Password,
		ProtocolID: serverReq.ProtocolID,
	}

	if err := h.service.Update(c.Request.Context(), srv); err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.ServerShortFromModel(*srv))
}

// Delete godoc
//
//	@Summary		Delete server
//	@Description	Deletes the server and unschedules its jobs.
//	@Tags			servers
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Server ID"
//	@Success		204	"No Content"
//	@Failure		401	{object}	response.Problem	"Missing or invalid token"
//	@Failure		404	{object}	response.Problem	"Server not found"
//	@Failure		500	{object}	response.Problem	"Internal server error"
//	@Router			/servers/{id} [delete]
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

// Execute godoc
//
//	@Summary		Execute RCON command
//	@Description	Sends a command to the server and returns its response.
//	@Tags			servers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Server ID"
//	@Param			command	body		request.ExecuteCommand	true	"Command"
//	@Success		200		{object}	response.ExecuteResult
//	@Failure		400		{object}	response.Problem	"Validation error"
//	@Failure		401		{object}	response.Problem	"Missing or invalid token"
//	@Failure		404		{object}	response.Problem	"Server not found"
//	@Failure		502		{object}	response.Problem	"Server did not respond over RCON"
//	@Failure		500		{object}	response.Problem	"Internal server error"
//	@Router			/servers/{id}/execute [post]
func (h *Handler) Execute(c *gin.Context) {
	var executeCommand request.ExecuteCommand
	if err := c.ShouldBindJSON(&executeCommand); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	var serverID path.ServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	res, err := h.service.Execute(c.Request.Context(), serverID.ID, executeCommand.Command)
	if err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.ExecuteResult{Output: res})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	servers := r.Group("/servers", middleware.Auth(h.cfg))
	{
		servers.GET("/:id", h.GetByID)
		servers.PUT("/:id", h.Update)
		servers.DELETE("/:id", h.Delete)
		servers.GET("", h.GetAll)
		servers.POST("", h.Create)
		servers.POST("/:id/execute", h.Execute)
	}
}
