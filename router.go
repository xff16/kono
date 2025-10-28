package bravka

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"slices"

	"go.uber.org/zap"

	"github.com/starwalkn/bravka/internal/logger"
)

type Router struct {
	dispatcher dispatcher
	aggregator aggregator
	Routes     []Route

	log *zap.Logger
}

func NewRouter(cfgs []RouteConfig) *Router {
	log := logger.Init(true)

	var (
		router = &Router{
			dispatcher: &defaultDispatcher{},
			aggregator: &defaultAggregator{},
			Routes:     nil,
			log:        log,
		}

		routes = make([]Route, 0, len(cfgs))
	)

	for _, cfg := range cfgs {
		// --- backends ---
		backends := make([]Backend, 0, len(cfg.Backends))
		for _, bcfg := range cfg.Backends {
			backend := Backend{
				URL:                 bcfg.URL,
				Method:              bcfg.Method,
				Timeout:             bcfg.Timeout,
				Headers:             bcfg.Headers,
				ForwardHeaders:      bcfg.ForwardHeaders,
				ForwardQueryStrings: bcfg.ForwardQueryStrings,
			}

			backends = append(backends, backend)
		}

		// --- plugins ---
		plugins := make([]Plugin, 0, len(cfg.Plugins))
		for _, pcfg := range cfg.Plugins {
			if slices.ContainsFunc(plugins, func(plugin Plugin) bool {
				return plugin.Name() == pcfg.Name
			}) {
				continue
			}

			soPlugin := loadPluginFromSO(pcfg.Path, pcfg.Config, log)
			if soPlugin == nil {
				log.Error(
					"cannot load plugin from .so",
					zap.String("name", pcfg.Name),
					zap.String("path", pcfg.Path),
				)
				continue
			}

			log.Info(
				"plugin initialized",
				zap.String("name", soPlugin.Name()),
				zap.String("route", cfg.Method+" "+cfg.Path),
			)

			plugins = append(plugins, soPlugin)
		}

		// --- middlewares ---
		middlewares := make([]Middleware, 0, len(cfg.Middlewares))
		for _, mcfg := range cfg.Middlewares {
			soMiddleware := loadMiddlewareFromSO(mcfg.Path, mcfg.Config, log)
			if soMiddleware == nil {
				log.Error(
					"cannot load middleware from .so",
					zap.String("name", mcfg.Name),
				)

				if !mcfg.CanFailOnLoad {
					panic("cannot load middleware from .so")
				}
				continue
			}

			log.Info(
				"middleware initialized",
				zap.String("name", soMiddleware.Name()),
				zap.String("route", cfg.Method+" "+cfg.Path),
			)

			middlewares = append(middlewares, soMiddleware)
		}

		// --- route ---
		route := Route{
			Path:        cfg.Path,
			Method:      cfg.Method,
			Backends:    backends,
			Aggregate:   cfg.Aggregate,
			Transform:   cfg.Transform,
			Plugins:     plugins,
			Middlewares: middlewares,
		}

		routes = append(routes, route)
	}

	router.Routes = routes

	return router
}

type Route struct {
	Path        string
	Method      string
	Backends    []Backend
	Aggregate   string
	Transform   string
	Plugins     []Plugin
	Middlewares []Middleware
}

type Backend struct {
	URL                 string
	Method              string
	Timeout             int64
	Headers             map[string]string
	ForwardHeaders      []string
	ForwardQueryStrings []string
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

	var routeHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// --- 1. Request-phase plugins ---
		for _, p := range rt.Plugins {
			if p.Type() != PluginTypeRequest {
				continue
			}

			pctx := &Context{
				Request: req,
				Route:   rt,
			}

			r.log.Debug("executing request plugin", zap.String("name", p.Name()))

			p.Execute(pctx)

			if pctx.Response != nil && pctx.Response.StatusCode == http.StatusTooManyRequests {
				r.log.Warn("too many requests", zap.String("request_uri", req.URL.RequestURI()))
				copyResponse(w, pctx.Response)

				return
			}
		}

		// --- 2. Backend dispatch ---
		responses := r.dispatcher.dispatch(rt, req)

		r.log.Debug("dispatched responses", zap.Any("responses", responses))

		aggregated := r.aggregator.aggregate(responses, os.Getenv("bravka_AGGREGATOR"))

		r.log.Debug("aggregated responses", zap.Any("aggregated", aggregated))

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

			r.log.Debug("executing response plugin", zap.String("name", p.Name()))

			p.Execute(pctx)
		}

		// --- 4. Write final output ---
		copyResponse(w, pctx.Response)
	})

	for i := len(rt.Middlewares) - 1; i >= 0; i-- {
		routeHandler = rt.Middlewares[i].Handler(routeHandler)
	}

	routeHandler.ServeHTTP(w, req)
}

func (r *Router) Match(req *http.Request) *Route {
	for _, route := range r.Routes {
		if route.Method != "" && route.Method != req.Method {
			continue
		}

		if route.Path != "" && req.URL.Path != route.Path {
			continue
		}

		return &route
	}

	return nil
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
