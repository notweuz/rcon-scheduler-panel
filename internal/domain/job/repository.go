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
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, job *model.Job) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return fmt.Errorf("create job: %w", ErrAlreadyExists)
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			return fmt.Errorf("create job: %w", ErrServerNotFound)
		}
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64, opts ...Option) (*model.Job, error) {
	var j model.Job
	if err := r.query(ctx, opts...).First(&j, id).Error; err != nil {
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
		return nil, fmt.Errorf("get all jobs of server %d: %w", serverID, err)
	}
	return jobs, nil
}

func (r *repository) GetAll(ctx context.Context, opts ...Option) ([]model.Job, error) {
	var jobs []model.Job
	if err := r.query(ctx, opts...).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("get all jobs: %w", err)
	}
	return jobs, nil
}

func (r *repository) Update(ctx context.Context, job *model.Job) error {
	res := r.db.WithContext(ctx).
		Model(&model.Job{ID: job.ID}).
		Select("name", "cron", "log_results", "steps").
		Updates(job)

	if err := res.Error; err != nil {
		return fmt.Errorf("update job %d: %w", job.ID, err)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update job %d: %w", job.ID, ErrNotFound)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&model.Job{}, id)
	if err := res.Error; err != nil {
		return fmt.Errorf("delete job %d: %w", id, err)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("delete job %d: %w", id, ErrNotFound)
	}
	return nil
}

// DRY for opts
func (r *repository) query(ctx context.Context, opts ...Option) *gorm.DB {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	db := r.db.WithContext(ctx)

	if o.IncludeServer {
		db = db.Preload("Server")
	}

	return db
}
