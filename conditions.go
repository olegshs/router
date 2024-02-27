package router

import (
	"github.com/julienschmidt/httprouter"
)

type conditions map[int]func(string) bool

func (c conditions) clone() conditions {
	clone := make(conditions, len(c))
	for k, r := range c {
		clone[k] = r
	}
	return clone
}

func (c conditions) match(params httprouter.Params) bool {
	for k, fn := range c {
		v := params[k].Value
		if !fn(v) {
			return false
		}
	}
	return true
}
