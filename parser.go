package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/olegshs/router/helpers"
)

var (
	parserRouteRegexp = regexp.MustCompile(`^(((GET|POST|PUT|PATCH|DELETE|OPTIONS)\b(,\s*)?)+)(\s+(.*))?$`)
	parserGroupRegexp = regexp.MustCompile(`^\(.*\)$`)

	regexpCache = regexpMap{}
)

type parser struct {
	router           *Router
	handlerByName    func(string) http.Handler
	middlewareByName func(string) MiddlewareFunc
}

func (p *parser) ParseMap(m map[string]interface{}) {
	keys := helpers.Map[string, interface{}](m).SortedKeys()
	for _, k := range keys {
		v := m[k]
		p.parseKeyValue(k, v)
	}
}

func (p *parser) parseKeyValue(k string, v interface{}) {
	if k[0] == '$' {
		p.parseKeyword(k, v)
		return
	}

	if a := parserRouteRegexp.FindStringSubmatch(k); len(a) > 0 {
		p.parseRoute(a, v)
	} else if m, ok := v.(map[string]interface{}); ok {
		if parserGroupRegexp.MatchString(k) {
			p.router.Group(func(r *Router) {
				r.ParseMap(m, p.handlerByName, p.middlewareByName)
			})
		} else {
			p.router.Prefix(k, func(r *Router) {
				r.ParseMap(m, p.handlerByName, p.middlewareByName)
			})
		}
	}
}

func (p *parser) parseKeyword(k string, v interface{}) {
	switch k {
	case "$where":
		conditions := v.(map[string]interface{})
		for k, v := range conditions {
			s := fmt.Sprint(v)
			r := regexpCache.Get(s)
			p.router.Where(k, r)
		}
	case "$use":
		switch t := v.(type) {
		case string:
			middleware := p.middlewareByName(t)
			if middleware != nil {
				p.router.Use(p.middlewareByName(t))
			}
		case []interface{}:
			for _, v := range t {
				name := fmt.Sprint(v)
				middleware := p.middlewareByName(name)
				if middleware != nil {
					p.router.Use(middleware)
				}
			}
		}
	}
}

func (p *parser) parseRoute(a []string, v interface{}) {
	var name string
	conditions := make(map[string]string)

	switch t := v.(type) {
	case string:
		name = t
	case map[string]interface{}:
		for k, v := range t {
			switch k {
			case "$name":
				name = fmt.Sprint(v)
			}
			if k[0] == '$' {
				continue
			}

			conditions[k] = fmt.Sprint(v)
		}
	}

	methods := helpers.Slice[string](strings.Split(a[1], ",")).Map(strings.TrimSpace)
	path := a[6]

	route := p.router.NewRoute(path, methods...).Name(name).Handle(p.handlerByName(name))
	for k, v := range conditions {
		r := regexp.MustCompile(v)
		route.Where(k, r)
	}
}
