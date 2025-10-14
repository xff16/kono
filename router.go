package kairyu

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/starwalkn/kairyu/internal/logger"
)

type Router struct {
	dispatcher dispatcher
	aggregator aggregator
	Routes     []Route

	log *zap.Logger
}

func NewRouter(cfgs []RouteConfig) *Router {
	routes := make([]Route, 0, len(cfgs))

	for _, cfg := range cfgs {
		// --- backends ---
		backends := make([]Backend, 0, len(cfg.Backends))
		for _, bcfg := range cfg.Backends {
			backend := Backend{
				URL:     bcfg.URL,
				Method:  bcfg.Method,
				Timeout: bcfg.Timeout,
			}

			backends = append(backends, backend)
		}

		// --- plugins ---
		plugins := make([]Plugin, 0, len(cfg.Plugins))
		for _, pcfg := range cfg.Plugins {
			p := createPlugin(pcfg.Name)
			if p == nil {
				log.Printf("plugin %s not found", pcfg.Name)
				continue
			}

			p.Init(pcfg.Config)
			plugins = append(plugins, p)
		}

		// --- route ---
		route := Route{
			Path:      cfg.Path,
			Method:    cfg.Method,
			Backends:  backends,
			Aggregate: cfg.Aggregate,
			Transform: cfg.Transform,
			Plugins:   plugins,
		}

		routes = append(routes, route)
	}

	return &Router{
		dispatcher: &defaultDispatcher{},
		aggregator: &defaultAggregator{},
		Routes:     routes,
		log:        logger.Init(true),
	}
}

type Route struct {
	Path      string
	Method    string
	Backends  []Backend
	Aggregate string
	Transform string
	Plugins   []Plugin
}

type Backend struct {
	URL     string
	Method  string
	Timeout int
}

/*
ServeHTTP is the incoming requests pipeline:

	ServeHTTP()
	 ├─ Match(req)
	 ├─ RunPlugins(PluginTypeRequest)
	 │    └─ могут вернуть Response (rate limit, auth)
	 ├─ Dispatch()
	 ├─ Aggregate()
	 ├─ RunPlugins(PluginTypeResponse)
	 └─ WriteResponse()
*/
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rt := r.Match(req)
	if rt == nil {
		r.log.Error("no route found", zap.String("request_uri", req.URL.RequestURI()))

		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	// --- 1. Request-phase plugins ---
	for _, p := range rt.Plugins {
		if p.Type() != PluginTypeRequest {
			continue
		}

		pctx := &Context{
			Request: req,
			Route:   rt,
		}

		p.Execute(pctx)

		if pctx.Response != nil && pctx.Response.StatusCode == http.StatusTooManyRequests {
			copyResponse(w, pctx.Response)
			return
		}
	}

	// --- 2. Backend dispatch ---
	responses := r.dispatcher.dispatch(rt, req)
	aggregated := r.aggregator.aggregate(responses, "merge")

	// --- 3. Response-phase plugins ---
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(aggregated)),
		Header:     make(http.Header),
	}

	pctx := &Context{
		Request:  req,
		Response: resp,
		Route:    rt,
	}

	for _, p := range rt.Plugins {
		if p.Type() != PluginTypeResponse {
			continue
		}

		p.Execute(pctx)
	}

	// --- 4. Write final output ---
	copyResponse(w, pctx.Response)
}

func (r *Router) Match(req *http.Request) *Route {
	return &r.Routes[0]
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		io.Copy(w, resp.Body)
	}
}
