package router

import (
	"regexp"
)

type regexpMap map[string]*regexp.Regexp

func (m regexpMap) Get(s string) *regexp.Regexp {
	r, ok := m[s]
	if !ok {
		r = regexp.MustCompile(s)
		m[s] = r
	}
	return r
}
