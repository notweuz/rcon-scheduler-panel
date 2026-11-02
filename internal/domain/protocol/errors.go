package protocol

import "errors"

var (
	ErrAlreadyExists = errors.New("protocol already exist")
	ErrNotFound      = errors.New("protocol not found")
)