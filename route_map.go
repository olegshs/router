package router

type routeMap map[string]map[string]*routeList

func (r routeMap) get(method string, pattern string) *routeList {
	if _, ok := r[method]; !ok {
		r[method] = make(map[string]*routeList)
	}

	if _, ok := r[method][pattern]; !ok {
		a := make(routeList, 0)
		r[method][pattern] = &a
	}

	return r[method][pattern]
}
