package bravka

import (
	"go.uber.org/zap"
)

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

type CorePlugin interface {
	Name() string
	Init(cfg map[string]interface{}) error // инициализация, чтение конфигурации
	Start() error                          // запуск/подключение к шлюзу
	Stop() error                           // остановка/отключение при reload/shutdown
}

//nolint:gochecknoglobals // non concurrently uses
var coreRegistry = make(map[string]func() CorePlugin)

func RegisterCorePlugin(name string, factory func() CorePlugin) {
	coreRegistry[name] = factory
}

func createCorePlugin(name string) CorePlugin {
	if f, ok := coreRegistry[name]; ok {
		return f()
	}
	return nil
}

func loadPluginFromSO(path string, cfg map[string]any, log *zap.Logger) Plugin {
	factory := loadSymbol[func() Plugin](path, "NewPlugin", log)

	plugin := factory()
	plugin.Init(cfg)

	return plugin
}
