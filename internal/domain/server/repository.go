package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, srv *model.Server) error {
	if err := r.db.WithContext(ctx).Create(srv).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create server: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("create server: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64, opts ...Option) (*model.Server, error) {
	var s model.Server
	if err := r.query(ctx, opts...).First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get server %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("get server %d: %w", id, err)
	}
	return &s, nil
}

func (r *repository) GetAll(ctx context.Context, opts ...Option) ([]model.Server, error) {
	var servers []model.Server
	if err := r.query(ctx, opts...).WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("get all servers: %w", err)
	}
	return servers, nil
}

func (r *repository) Delete(ctx context.Context, id int64, opts ...Option) (*model.Server, error) {
	var server model.Server

	res := r.query(ctx, opts...).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "address"}, {Name: "port"}}}).
		Delete(&server, id)

	if err := res.Error; err != nil {
		return nil, fmt.Errorf("delete server %d: %w", id, err)
	}

	// At first, I though that repository shouldn't return an error, if no fields are affected. But it's kinda needed for http layer
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("delete server %d: %w", id, ErrNotFound)
	}
	return &server, nil
}

// DRY for opts
func (r *repository) query(ctx context.Context, opts ...Option) *gorm.DB {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	db := r.db.WithContext(ctx)

	if o.IncludeProtocol {
		db = db.Preload("Protocol")
	}

	return db
}
