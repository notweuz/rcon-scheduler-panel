package job

import "errors"

var (
	ErrAlreadyExists = errors.New("job already exist")
	ErrNotFound      = errors.New("job not found")
)