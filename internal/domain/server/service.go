package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
)

type Repository interface {
	Create(ctx context.Context, srv *model.Server) error
	GetByID(ctx context.Context, id int64) (*model.Server, error)
	GetAll(ctx context.Context) ([]model.Server, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, srv *model.Server) error {
	if err := s.repo.Create(ctx, srv); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("server already exist").WithErr(err)
		}
		return fmt.Errorf("create server: %w", err)
	}
	return nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.Server, error) {
	srv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return nil, fmt.Errorf("get server: %w", err)
	}
	return srv, nil
}

func (s *service) GetAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all servers: %w", err)
	}

	if servers == nil {
		servers = []model.Server{}
	}

	return servers, nil
}