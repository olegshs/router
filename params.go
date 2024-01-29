package router

import (
	"context"
	"net/http"
)

type Params []Param

type Param struct {
	Key   string
	Value string
}

type paramsKeyType struct{}

var paramsKey = paramsKeyType{}

// ParamsFromRequest retrieves a structure with named parameters from an HTTP request.
func ParamsFromRequest(r *http.Request) Params {
	params, _ := r.Context().Value(paramsKey).(Params)
	return params
}

// ByName returns the value of a parameter by its name.
func (params Params) ByName(name string) string {
	for _, p := range params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Values returns an array of strings with the values of all parameters.
func (params Params) Values() []string {
	a := make([]string, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// InterfaceValues returns an array of empty interfaces with the values of all parameters.
func (params Params) InterfaceValues() []interface{} {
	a := make([]interface{}, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// Map returns a map of all parameters. The values are strings.
func (params Params) Map() map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

// InterfaceMap returns a map of all parameters. The values are empty interfaces.
func (params Params) InterfaceMap() map[string]interface{} {
	m := make(map[string]interface{}, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

func (params Params) toRequest(r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, paramsKey, params)
	*r = *r.WithContext(ctx)
}
