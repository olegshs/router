package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/olegshs/router/helpers"
)

type Route struct {
	router          *Router
	methods         []string
	pattern         pattern
	paramNames      helpers.Slice[string]
	paramNamesMatch [][]string
	conditions      conditions
	handler         http.Handler
}

// Name sets a name of the route.
func (route *Route) Name(name string) *Route {
	route.router.routeByName[name] = route
	return route
}

// Where sets a regular expression for validating a named parameter.
func (route *Route) Where(param string, regex *regexp.Regexp) *Route {
	i := route.paramNames.IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	route.conditions[i] = func(v string) bool {
		return regex.MatchString(v)
	}
	return route
}

// WhereFunc sets a function for validating a named parameter.
func (route *Route) WhereFunc(param string, matchFunc func(string) bool) *Route {
	i := route.paramNames.IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	route.conditions[i] = matchFunc
	return route
}

// Handle sets a handler for the route.
func (route *Route) Handle(handler http.Handler) *Route {
	route.handler = handler
	return route
}

// HandleFunc sets a handler for the route.
func (route *Route) HandleFunc(handlerFunc http.HandlerFunc) *Route {
	return route.Handle(handlerFunc)
}

// Url generates a URL for the route.
func (route *Route) Url(params ...interface{}) (string, error) {
	nParams := len(params)
	nMatch := len(route.paramNamesMatch)
	if nParams < nMatch {
		err := fmt.Errorf("%w (%d < %d)",
			ErrNotEnoughParameters, nParams, nMatch,
		)
		return "", err
	}

	u := string(route.pattern)

	for i, v := range params {
		s := fmt.Sprint(v)

		if (route.conditions[i] != nil) && !route.conditions[i](s) {
			err := fmt.Errorf("%w: %s not match the conditions",
				ErrInvalidParameter, strconv.Quote(s),
			)
			return "", err
		}

		if i < nMatch {
			m := route.paramNamesMatch[i]
			u = strings.ReplaceAll(u, m[0], s)
		} else {
			u += "/" + s
		}
	}

	return u, nil
}

func (route *Route) namedParams(params httprouter.Params) Params {
	n := len(params)
	if n == 0 {
		return nil
	}

	named := make(Params, n)
	for i, param := range params {
		named[i] = Param{
			Key:   route.paramNames[i],
			Value: param.Value,
		}
	}

	return named
}
