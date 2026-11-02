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

func (r *repository) Create(ctx context.Context, protocol *model.Protocol) error {
	if err := r.db.WithContext(ctx).Create(protocol).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create protocol: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("create protocol: %w", err)
	}
	return nil
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

func (r *repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&model.Protocol{}, id)
	
	if err := res.Error; err != nil {
		return fmt.Errorf("delete protocol %d: %w", id, err)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("delete protocol %d: %w", id, ErrNotFound)
	}
	return nil
}
