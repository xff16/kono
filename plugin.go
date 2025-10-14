package kairyu

type Factory func() Plugin

var registry = make(map[string]Factory)

func Register(name string, factory Factory) {
	registry[name] = factory
}

func createPlugin(name string) Plugin {
	if f, ok := registry[name]; ok {
		return f()
	}
	return nil
}

type Plugin interface {
	Name() string
	Init(cfg map[string]any)
	Type() PluginType
	Execute(ctx *Context)
}

type PluginType int

const (
	PluginTypeRequest  = iota // JWT, rate limit, logging
	PluginTypeResponse        // Transform, mask, log
)

type BasePlugin struct {
	pluginType PluginType
}

func (bp *BasePlugin) SetType(t PluginType) { bp.pluginType = t }
func (bp *BasePlugin) Type() PluginType     { return bp.pluginType }
