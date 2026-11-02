package protocol

import (
	"context"
	"errors"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
)

type Repository interface {
	Create(ctx context.Context, protocol *model.Protocol) error
	GetByID(ctx context.Context, id int64) (*model.Protocol, error)
	GetAll(ctx context.Context) ([]model.Protocol, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, protocol *model.Protocol) error {
	if err := s.repo.Create(ctx, protocol); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("protocol already exist").WithErr(err)
		}
		return err
	}
	return nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.Protocol, error) {
	protocol, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("protocol not found").WithErr(err)
		}
		return nil, err
	}
	return protocol, nil
}

func (s *service) GetAll(ctx context.Context) ([]model.Protocol, error) {
	protocols, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if protocols == nil {
		protocols = []model.Protocol{}
	}

	return protocols, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return errs.ErrNotFound.WithMessage("protocol not found").WithErr(err)
		}
		return err
	}
	return nil
}