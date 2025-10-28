package bravka

import (
	"encoding/json"
	"log"
	"os"
)

type GatewayConfig struct {
	Schema    string             `json:"schema"`
	Name      string             `json:"name"`
	Version   string             `json:"version"`
	Server    ServerConfig       `json:"server"`
	Dashboard DashboardConfig    `json:"dashboard"`
	Plugins   []CorePluginConfig `json:"plugins"`
	Routes    []RouteConfig      `json:"routes"`
}

type ServerConfig struct {
	Port    int `json:"port"`
	Timeout int `json:"timeout"`
}
type DashboardConfig struct {
	Enable bool `json:"enable"`
	Port   int  `json:"port"`
}

type RouteConfig struct {
	Path        string             `json:"path"`
	Method      string             `json:"method"`
	Plugins     []PluginConfig     `json:"plugins"`
	Middlewares []MiddlewareConfig `json:"middlewares"`
	Backends    []BackendConfig    `json:"backends"`
	Aggregate   string             `json:"aggregate"`
	Transform   string             `json:"transform"`
}

type BackendConfig struct {
	URL                 string            `json:"url"`
	Method              string            `json:"method"`
	Timeout             int64             `json:"timeout"`
	Headers             map[string]string `json:"headers"`
	ForwardHeaders      []string          `json:"forward_headers"`
	ForwardQueryStrings []string          `json:"forward_query_strings"`
}

type PluginConfig struct {
	Name   string                 `json:"name"`
	Path   string                 `json:"path,omitempty"`
	Config map[string]interface{} `json:"config"`
}

type MiddlewareConfig struct {
	Name          string                 `json:"name"`
	Path          string                 `json:"path,omitempty"`
	Config        map[string]interface{} `json:"config"`
	CanFailOnLoad bool                   `json:"can_fail_on_load"`
}

type CorePluginConfig struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

func LoadConfig(path string) GatewayConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("failed to read config file:", err)
	}

	var cfg GatewayConfig
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("failed to parse json:", err)
	}

	for _, pcfg := range cfg.Plugins {
		p := createCorePlugin(pcfg.Name)
		if p == nil {
			log.Println("failed to create core plugin:", pcfg.Name)
			continue
		}

		if err = p.Init(pcfg.Config); err != nil {
			log.Fatal("failed to init core plugin:", err)
		}

		if err = p.Start(); err != nil {
			log.Fatal("failed to start core plugin:", err)
		}
	}

	return cfg
}
