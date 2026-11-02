package protocol

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id int64) (*model.Protocol, error) {
	var p model.Protocol
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get protocol %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("get protocol %d: %w", id, err)
	}
	return &p, nil
}

func (r *repository) GetAll(ctx context.Context) ([]model.Protocol, error) {
	var protocols []model.Protocol
	if err := r.db.WithContext(ctx).Find(&protocols).Error; err != nil {
		return nil, fmt.Errorf("get all protocols: %w", err)
	}
	return protocols, nil
}