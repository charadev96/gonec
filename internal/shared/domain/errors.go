package domain

import (
	"errors"
)

var ErrNotExist = errors.New("entry does not exist in repository")
