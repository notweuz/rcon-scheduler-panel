package server

import "errors"

var (
	ErrAlreadyExists = errors.New("server already exist")
)