// Package router implements an HTTP request router.
//
// It serves as a wrapper for the package [github.com/julienschmidt/httprouter], extending its capabilities.
//
// Additional features include:
//   - generating URLs for named routes
//   - route prefixes and groups
//   - middleware functions
//   - validation of named parameters using regular expressions
package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	prefix      pattern
	conditions  conditions
	middleware  middlewareList
	routes      routeMap
	routeByName map[string]*Route
	r           *httprouter.Router
}

// New creates a new instance of the router.
func New() *Router {
	router := new(Router)
	router.prefix = ""
	router.conditions = make(conditions)
	router.middleware = make(middlewareList, 0)
	router.routes = make(routeMap)
	router.routeByName = make(map[string]*Route)
	router.r = httprouter.New()
	router.r.NotFound = http.NotFoundHandler()

	return router
}

// ParseMap adds routes defined in a map with a special structure (see example).
func (router *Router) ParseMap(
	m map[string]interface{},
	handlerByName func(string) http.Handler,
	middlewareByName func(string) MiddlewareFunc,
) {
	p := &parser{
		router:           router,
		handlerByName:    handlerByName,
		middlewareByName: middlewareByName,
	}
	p.ParseMap(m)
}

// Group adds a group of routes.
// Middleware functions can be specified for the group.
func (router *Router) Group(f func(*Router)) {
	f(router.clone())
}

// Prefix adds a group of routes with a specified prefix.
// The prefix can contain named parameters.
func (router *Router) Prefix(path string, f func(*Router)) {
	sub := router.clone()
	sub.prefix = router.prefix + pattern(path)

	f(sub)
}

// Use adds middleware functions that will be used by the router or by a group of routes.
func (router *Router) Use(middleware ...MiddlewareFunc) {
	router.middleware = append(router.middleware, middleware...)
}

// Where sets a regular expression for validating the named parameter specified in a prefix.
func (router *Router) Where(param string, regexp *regexp.Regexp) {
	router.WhereFunc(param, func(v string) bool {
		return regexp.MatchString(v)
	})
}

// WhereFunc sets a function for validating the named parameter specified in a prefix.
func (router *Router) WhereFunc(param string, matchFunc func(string) bool) {
	i := router.prefix.paramNames().IndexOf(param)
	if i < 0 {
		panic("unknown parameter: " + param)
	}

	router.conditions[i] = matchFunc
}

// Get creates and returns a route for handling GET requests.
func (router *Router) Get(path string) *Route {
	return router.NewRoute(path, http.MethodGet)
}

// Post creates and returns a route for handling POST requests.
func (router *Router) Post(path string) *Route {
	return router.NewRoute(path, http.MethodPost)
}

// Put creates and returns a route for handling PUT requests.
func (router *Router) Put(path string) *Route {
	return router.NewRoute(path, http.MethodPut)
}

// Patch creates and returns a route for handling PATCH requests.
func (router *Router) Patch(path string) *Route {
	return router.NewRoute(path, http.MethodPatch)
}

// Delete creates and returns a route for handling DELETE requests.
func (router *Router) Delete(path string) *Route {
	return router.NewRoute(path, http.MethodDelete)
}

// Options creates and returns a route for handling OPTIONS requests.
func (router *Router) Options(path string) *Route {
	return router.NewRoute(path, http.MethodOptions)
}

// NewRoute creates and returns a route for handling requests sent with the specified methods.
func (router *Router) NewRoute(path string, methods ...string) *Route {
	route := new(Route)
	route.router = router
	route.methods = methods
	route.pattern = router.prefix + pattern(path)
	route.paramNames = route.pattern.paramNames()
	route.paramNamesMatch = route.pattern.paramNamesMatch()
	route.conditions = router.conditions.clone()

	router.addRoute(route)

	return route
}

// Url generates a URL for a named route.
func (router *Router) Url(name string, params ...interface{}) (string, error) {
	route, ok := router.routeByName[name]
	if !ok {
		return "", fmt.Errorf("%s: %w", name, ErrRouteNotFound)
	}

	u, err := route.Url(params...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}

	return u, nil
}

// HandleNotFound sets a handler that is called when a route is not found.
func (router *Router) HandleNotFound(handler http.Handler) {
	router.r.NotFound = router.middleware.wrap(handler)
}

// HandleMethodNotAllowed sets a handler that is called when the route is found,
// but the request method is not supported.
func (router *Router) HandleMethodNotAllowed(handler http.Handler) {
	router.r.MethodNotAllowed = router.middleware.wrap(handler)
}

// HandlePanic sets a panic handler for the router.
// The handler receives http.ResponseWriter, *http.Request,
// and the value returned by the recover function.
func (router *Router) HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	router.r.PanicHandler = handler
}

// ServeHTTP implements the http.Handler interface.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.r.ServeHTTP(w, r)
}

func (router *Router) clone() *Router {
	clone := new(Router)
	clone.prefix = router.prefix
	clone.conditions = router.conditions.clone()
	clone.middleware = router.middleware.clone()
	clone.routes = router.routes
	clone.routeByName = router.routeByName
	clone.r = router.r

	return clone
}

func (router *Router) addRoute(route *Route) {
	p := route.pattern.httpRouterString()

	for _, method := range route.methods {
		a := router.routes.get(method, p)

		if len(*a) == 0 {
			h := router.newHandler(a)
			router.r.Handler(method, p, h)
		}

		*a = append(*a, route)
	}
}

func (router *Router) newHandler(routes *routeList) http.Handler {
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := httprouter.ParamsFromContext(r.Context())
		for i, param := range params {
			params[i].Value = strings.Trim(param.Value, "/")
		}

		route := routes.match(params)
		if route == nil {
			router.r.NotFound.ServeHTTP(w, r)
			return
		}

		namedParams := route.namedParams(params)
		if len(namedParams) > 0 {
			namedParams.toRequest(r)
		}

		route.handler.ServeHTTP(w, r)
	})

	handler = router.middleware.wrap(handler)

	return handler
}
