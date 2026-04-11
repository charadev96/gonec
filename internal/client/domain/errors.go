package domain

import (
	"errors"
)

var (
	ErrConn       = errors.New("already connected")
	ErrNoConn     = errors.New("no active connection")
	ErrLoggedIn   = errors.New("already logged in")
	ErrNoLoggedIn = errors.New("not logged in")
)
