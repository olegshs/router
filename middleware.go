package router

import (
	"net/http"
)

type MiddlewareFunc func(http.Handler) http.Handler

type middlewareList []MiddlewareFunc

func (middleware middlewareList) clone() middlewareList {
	clone := make(middlewareList, len(middleware))
	copy(clone, middleware)
	return clone
}

func (middleware middlewareList) wrap(handler http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
