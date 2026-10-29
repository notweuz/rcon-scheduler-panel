package server

import "errors"

var (
	ErrAlreadyExists = errors.New("server already exist")
	ErrNotFound      = errors.New("server not found")
)
