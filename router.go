package main

import "net/http"

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Router struct {
	routes map[string]map[string]HandlerFunc
	prefix string
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
		prefix: "",
	}
}

func (r *Router) Group(p string, fn func(*Router)) {
	pr := r.prefix
	r.prefix = pr + p
	fn(r)
	r.prefix = pr
}

func (r *Router) Handle(method string, path string, h HandlerFunc) {
	full := r.prefix + path
	if r.routes[full] == nil {
		r.routes[full] = make(map[string]HandlerFunc)
	}
	r.routes[full][method] = h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m := req.Method
	p := req.URL.Path
	if methods, ok := r.routes[p]; ok {
		if h, ok := methods[m]; ok {
			h(w, req)
			return
		}
	}
	http.NotFound(w, req)
}
