package router

import (
	"errors"
)

var (
	ErrRouteNotFound       = errors.New("route not found")
	ErrNotEnoughParameters = errors.New("not enough parameters")
	ErrInvalidParameter    = errors.New("invalid parameter")
)
