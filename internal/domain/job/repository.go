package job

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
	return &repository{}
}

func (r *repository) Create(ctx context.Context, job *model.Job) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("create job: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*model.Job, error) {
	var j model.Job
	if err := r.db.WithContext(ctx).First(&j, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get job %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("get job %d: %w", id, err)
	}
	return &j, nil
}

func (r *repository) GetByServerID(ctx context.Context, serverID int64) ([]model.Job, error) {
	var jobs []model.Job
	if err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("get all jobs: %w", err)
	}
	return jobs, nil
}