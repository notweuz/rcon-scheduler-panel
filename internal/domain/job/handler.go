package job

import (
	"context"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/model"
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

