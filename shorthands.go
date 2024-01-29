package router

import (
	"net/http"
)

var (
	defaultRouter *Router
)

// DefaultRouter returns the default router instance.
func DefaultRouter() *Router {
	if defaultRouter == nil {
		defaultRouter = New()
	}
	return defaultRouter
}

// ParseMap adds routes defined in a map with a special structure (see example).
func ParseMap(
	m map[string]interface{},
	handlerByName func(string) http.Handler,
	middlewareByName func(string) MiddlewareFunc,
) {
	DefaultRouter().ParseMap(m, handlerByName, middlewareByName)
}

// Group adds a group of routes.
// Middleware functions can be specified for the group.
func Group(f func(*Router)) {
	DefaultRouter().Group(f)
}

// Prefix adds a group of routes with a specified prefix.
// The prefix can contain named parameters.
func Prefix(path string, f func(*Router)) {
	DefaultRouter().Prefix(path, f)
}

// Use adds middleware functions that will be used by the router.
func Use(middleware ...MiddlewareFunc) {
	DefaultRouter().Use(middleware...)
}

// Get creates and returns a route for handling GET requests.
func Get(path string) *Route {
	return DefaultRouter().Get(path)
}

// Post creates and returns a route for handling POST requests.
func Post(path string) *Route {
	return DefaultRouter().Post(path)
}

// Put creates and returns a route for handling PUT requests.
func Put(path string) *Route {
	return DefaultRouter().Put(path)
}

// Patch creates and returns a route for handling PATCH requests.
func Patch(path string) *Route {
	return DefaultRouter().Patch(path)
}

// Delete creates and returns a route for handling DELETE requests.
func Delete(path string) *Route {
	return DefaultRouter().Delete(path)
}

// Options creates and returns a route for handling OPTIONS requests.
func Options(path string) *Route {
	return DefaultRouter().Options(path)
}

// NewRoute creates and returns a route for handling requests sent with the specified methods.
func NewRoute(path string, methods ...string) *Route {
	return DefaultRouter().NewRoute(path, methods...)
}

// Url generates a URL for a named route.
func Url(name string, params ...interface{}) (string, error) {
	return DefaultRouter().Url(name, params...)
}

// HandleNotFound sets a handler that is called when a route is not found.
func HandleNotFound(handler http.Handler) {
	DefaultRouter().HandleNotFound(handler)
}

// HandleMethodNotAllowed sets a handler that is called when the route is found,
// but the request method is not supported.
func HandleMethodNotAllowed(handler http.Handler) {
	DefaultRouter().HandleMethodNotAllowed(handler)
}

// HandlePanic sets a panic handler for the router.
// The handler receives http.ResponseWriter, *http.Request,
// and the value returned by the recover function.
func HandlePanic(handler func(http.ResponseWriter, *http.Request, interface{})) {
	DefaultRouter().HandlePanic(handler)
}
