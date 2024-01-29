package router

import (
	"regexp"

	"github.com/julienschmidt/httprouter"
)

type conditions map[int]*regexp.Regexp

func (c conditions) clone() conditions {
	clone := make(conditions, len(c))
	for k, r := range c {
		clone[k] = r
	}
	return clone
}

func (c conditions) match(params httprouter.Params) bool {
	for k, r := range c {
		v := params[k].Value
		if !r.MatchString(v) {
			return false
		}
	}
	return true
}
