package job

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
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id int64) (*model.Job, error)
	GetByServerID(ctx context.Context, serverID int64) ([]model.Job, error)
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

func (h *Handler) GetByID(c *gin.Context) {
	var jobID path.JobID
	if err := c.ShouldBindUri(&jobID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	job, err := h.service.GetByID(c.Request.Context(), jobID.ID)
	if err != nil {
		respond.Error(c, err)
		return
	}
	respond.OK(c, response.JobFromModel(*job))
}

func (h *Handler) GetByServerID(c *gin.Context) {
	var serverID path.JobServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	jobs, err := h.service.GetByServerID(c.Request.Context(), serverID.ServerID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	jobsDTOs := make([]response.Job, len(jobs))
	for i, job := range jobs {
		jobsDTOs[i] = response.JobFromModel(job)
	}
	respond.OK(c, jobsDTOs)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/servers/:id/jobs", middleware.Auth(h.cfg), )
	jobs := r.Group("/jobs", middleware.Auth(h.cfg))
	{
		jobs.GET("/:id", h.GetByID)
	}
}