package router

import (
	"context"
	"net/http"
	"strings"
)

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

type router struct {
	routes []route
}

func New(ctx context.Context) *router {
	router := &router{
		routes: []route{},
	}

	return router
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		params, ok := match(route.pattern, req.URL.Path)
		if ok && req.Method == route.method {
			// Store params in context
			ctx := context.WithValue(req.Context(), "params", params)
			route.handler(w, req.WithContext(ctx))
			return
		}
	}
	http.NotFound(w, req)
}

func (r *router) AddRoute(method, path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, route{
		method:  method,
		pattern: path,
		handler: handler,
	})
}

func match(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return nil, false
	}
	params := make(map[string]string)
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}") {
			key := patternParts[i][1 : len(patternParts[i])-1]
			params[key] = pathParts[i]
		} else if patternParts[i] != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}
