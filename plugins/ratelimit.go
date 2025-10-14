package plugins

import (
	"net/http"
	"sync"
	"time"

	"github.com/starwalkn/kairyu"
)

type RateLimitPlugin struct {
	kairyu.BasePlugin

	limit   int
	window  time.Duration
	mu      sync.Mutex
	counter map[string]int
	reset   map[string]time.Time
}

func (p *RateLimitPlugin) Name() string            { return "rate_limit" }
func (p *RateLimitPlugin) Type() kairyu.PluginType { return kairyu.PluginTypeRequest }

func (p *RateLimitPlugin) Init(cfg map[string]any) {
	p.limit = intFrom(cfg, "limit", 60)                                // лимит на минуту
	p.window = time.Duration(intFrom(cfg, "window", 60)) * time.Second // окно
	p.counter = make(map[string]int)
	p.reset = make(map[string]time.Time)
}

func (p *RateLimitPlugin) Execute(ctx *kairyu.Context) {
	ip := ctx.Request.RemoteAddr

	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if exp, ok := p.reset[ip]; !ok || now.After(exp) {
		p.reset[ip] = now.Add(p.window)
		p.counter[ip] = 0
	}

	p.counter[ip]++
	if p.counter[ip] > p.limit {
		ctx.Response = &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       http.NoBody,
		}
	}
}

func intFrom(cfg map[string]any, key string, def int) int {
	if v, ok := cfg[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return def
}
