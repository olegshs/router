package router

import (
	"github.com/julienschmidt/httprouter"
)

type routeList []*Route

func (routes *routeList) match(params httprouter.Params) *Route {
	for _, route := range *routes {
		if route.handler == nil {
			continue
		}
		if route.conditions.match(params) {
			return route
		}
	}
	return nil
}
