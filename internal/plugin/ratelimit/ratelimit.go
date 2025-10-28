package ratelimit

import (
	"sync"
	"time"

	"github.com/starwalkn/bravka"
)

func init() {
	bravka.RegisterCorePlugin("ratelimit", NewPlugin)
}

type Plugin struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	counter map[string]int
	reset   map[string]time.Time
	stopCh  chan struct{}
}

func NewPlugin() bravka.CorePlugin {
	return &Plugin{}
}

func (p *Plugin) Name() string { return "ratelimit" }

func (p *Plugin) Init(cfg map[string]interface{}) error {
	p.limit = intFrom(cfg, "limit", 60)                                // лимит на минуту
	p.window = time.Duration(intFrom(cfg, "window", 60)) * time.Second // окно
	p.counter = make(map[string]int)
	p.reset = make(map[string]time.Time)
	p.stopCh = make(chan struct{})

	return nil
}

func (p *Plugin) Start() error {
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.cleanup()
			case <-p.stopCh:
				return
			}
		}
	}()
	return nil
}

func (p *Plugin) Stop() error {
	close(p.stopCh)
	return nil
}

func (p *Plugin) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for ip, reset := range p.reset {
		if now.After(reset) {
			delete(p.reset, ip)
			delete(p.counter, ip)
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
