package server

import (
	"context"
	"rcon-scheduler-panel/internal/model"
)

type Service interface {
	Create(ctx context.Context, srv *model.Server) error
	GetByID(ctx context.Context, id int64) (*model.Server, error)
	GetAll(ctx context.Context) ([]model.Server, error)
}
