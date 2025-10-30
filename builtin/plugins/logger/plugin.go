package main

import (
	"fmt"

	"github.com/starwalkn/bravka"
)

type Plugin struct {
	bravka.BasePlugin
}

func NewPlugin() bravka.Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "logger"
}

func (p *Plugin) Type() bravka.PluginType {
	return bravka.PluginTypeRequest
}

//nolint:revive // not implemented yet
func (p *Plugin) Init(cfg map[string]any) {}

//nolint:forbidigo // stub method
func (p *Plugin) Execute(ctx *bravka.Context) {
	fmt.Printf("[logger] %s %s\n", ctx.Request.Method, ctx.Request.URL.Path)
}
