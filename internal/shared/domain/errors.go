package domain

import (
	"errors"
)

var (
	ErrNotExist = errors.New("resource does not exist")
	ErrExist    = errors.New("resource already exists")
)
