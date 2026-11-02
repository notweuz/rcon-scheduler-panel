package job

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
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id int64) (*model.Job, error)
	GetByServerID(ctx context.Context, serverID int64) ([]model.Job, error)
	GetAll(ctx context.Context) ([]model.Job, error)
	Update(ctx context.Context, job *model.Job) error
	Delete(ctx context.Context, id int64) error
	Resume(ctx context.Context) error
	ScheduleServer(ctx context.Context, serverID int64) error
	UnscheduleServer(serverID int64) error
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
	var serverID path.ServerID
	if err := c.ShouldBindUri(&serverID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	var jobReq request.CreateJob
	if err := c.ShouldBindJSON(&jobReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	job := &model.Job{
		ServerID:   serverID.ID,
		Name:       jobReq.Name,
		Cron:       jobReq.Cron,
		LogResults: jobReq.LogResults,
		Steps:      stepsFromRequest(jobReq.Steps),
	}

	if err := h.service.Create(c.Request.Context(), job); err != nil {
		respond.Error(c, err)
		return
	}

	respond.Created(c, response.JobFromModel(*job))
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

func (h *Handler) Update(c *gin.Context) {
	var jobID path.JobID
	if err := c.ShouldBindUri(&jobID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	var jobReq request.UpdateJob
	if err := c.ShouldBindJSON(&jobReq); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	job := &model.Job{
		ID:         jobID.ID,
		Name:       jobReq.Name,
		Cron:       jobReq.Cron,
		LogResults: jobReq.LogResults,
		Steps:      stepsFromRequest(jobReq.Steps),
	}

	if err := h.service.Update(c.Request.Context(), job); err != nil {
		respond.Error(c, err)
		return
	}

	respond.OK(c, response.JobFromModel(*job))
}

func (h *Handler) Delete(c *gin.Context) {
	var jobID path.JobID
	if err := c.ShouldBindUri(&jobID); err != nil {
		respond.Error(c, validation.FromBindError(err))
		return
	}

	if err := h.service.Delete(c.Request.Context(), jobID.ID); err != nil {
		respond.Error(c, err)
		return
	}

	respond.NoContent(c)
}

func stepsFromRequest(steps []request.Step) []model.Step {
	result := make([]model.Step, len(steps))
	for i, s := range steps {
		result[i] = model.Step{
			Command: s.Command,
			Wait:    s.Wait,
		}
	}
	return result
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/servers/:id/jobs", middleware.Auth(h.cfg), h.GetByServerID)
	r.POST("/servers/:id/jobs", middleware.Auth(h.cfg), h.Create)
	jobs := r.Group("/jobs", middleware.Auth(h.cfg))
	{
		jobs.GET("/:id", h.GetByID)
		jobs.PUT("/:id", h.Update)
		jobs.DELETE("/:id", h.Delete)
	}
}
