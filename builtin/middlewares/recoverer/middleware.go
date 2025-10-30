package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"go.uber.org/zap"

	"github.com/starwalkn/bravka"
	"github.com/starwalkn/bravka/internal/logger"
)

type Middleware struct {
	enabled      bool
	includeStack bool
	log          *zap.Logger
}

func NewMiddleware() bravka.Middleware {
	return &Middleware{}
}

func (p *Middleware) Name() string {
	return "recoverer"
}

func (p *Middleware) Init(cfg map[string]interface{}) error {
	if val, ok := cfg["enabled"].(bool); ok {
		p.enabled = val
	}

	p.log = logger.New(true)

	return nil
}

func (p *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				now := time.Now()
				msg := fmt.Sprintf("panic recovered: %v", rec)

				log := p.log.With(
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Time("time", now),
				)

				if p.includeStack {
					log.Error(msg, zap.ByteString("stack", debug.Stack()))
				} else {
					log.Error(msg)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "internal server error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
