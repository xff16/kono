package tokka

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type GatewayConfig struct {
	Schema      string             `json:"schema" yaml:"schema" toml:"schema"`
	Name        string             `json:"name" yaml:"name" toml:"name"`
	Version     string             `json:"version" yaml:"version" toml:"version"`
	Debug       bool               `json:"debug" yaml:"debug" toml:"debug"`
	Server      ServerConfig       `json:"server" yaml:"server" toml:"server"`
	Dashboard   DashboardConfig    `json:"dashboard" yaml:"dashboard" toml:"dashboard"`
	Plugins     []CorePluginConfig `json:"plugins" yaml:"plugins" toml:"plugins"`
	Middlewares []MiddlewareConfig `json:"middlewares" yaml:"middlewares" toml:"middlewares"`
	Routes      []RouteConfig      `json:"routes" yaml:"routes" toml:"routes"`
}

type ServerConfig struct {
	Port          int  `json:"port" yaml:"port" toml:"port"`
	Timeout       int  `json:"timeout" yaml:"timeout" toml:"timeout"`
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics" toml:"enable_metrics"`
}

type DashboardConfig struct {
	Enable  bool `json:"enable" yaml:"enable" toml:"enable"`
	Port    int  `json:"port" yaml:"port" toml:"port"`
	Timeout int  `json:"timeout" yaml:"timeout" toml:"timeout"`
}

type RouteConfig struct {
	Path                string             `json:"path" yaml:"path" toml:"path"`
	Method              string             `json:"method" yaml:"method" toml:"method"`
	Plugins             []PluginConfig     `json:"plugins" yaml:"plugins" toml:"plugins"`
	Middlewares         []MiddlewareConfig `json:"middlewares" yaml:"middlewares" toml:"middlewares"`
	Backends            []BackendConfig    `json:"backends" yaml:"backends" toml:"backends"`
	Aggregate           string             `json:"aggregate" yaml:"aggregate" toml:"aggregate"`
	Transform           string             `json:"transform" yaml:"transform" toml:"transform"`
	AllowPartialResults bool               `json:"allow_partial_results" yaml:"allow_partial_results" toml:"allow_partial_results"`
}

type BackendConfig struct {
	URL                 string            `json:"url" yaml:"url" toml:"url"`
	Method              string            `json:"method" yaml:"method" toml:"method"`
	Timeout             int64             `json:"timeout" yaml:"timeout" toml:"timeout"`
	Headers             map[string]string `json:"headers" yaml:"headers" toml:"headers"`
	ForwardHeaders      []string          `json:"forward_headers" yaml:"forward_headers" toml:"forward_headers"`
	ForwardQueryStrings []string          `json:"forward_query_strings" yaml:"forward_query_strings" toml:"forward_query_strings"`
}

type PluginConfig struct {
	Name   string         `json:"name" yaml:"name" toml:"name"`
	Path   string         `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	Config map[string]any `json:"config" yaml:"config" toml:"config"`
}

type MiddlewareConfig struct {
	Name          string         `json:"name" yaml:"name" toml:"name"`
	Path          string         `json:"path,omitempty" yaml:"path,omitempty" toml:"path,omitempty"`
	Config        map[string]any `json:"config" yaml:"config" toml:"config"`
	CanFailOnLoad bool           `json:"can_fail_on_load" yaml:"can_fail_on_load" toml:"can_fail_on_load"`
	Override      bool           `json:"override" yaml:"override" toml:"override"`
}

type CorePluginConfig struct {
	Name   string         `json:"name" yaml:"name" toml:"name"`
	Config map[string]any `json:"config" yaml:"config" toml:"config"`
}

func LoadConfig(path string) GatewayConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("failed to read config file:", err)
	}

	var cfg GatewayConfig

	switch filepath.Ext(path) {
	case ".json":
		if err = json.Unmarshal(data, &cfg); err != nil {
			log.Fatal("failed to parse json:", err)
		}
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatal("failed to parse yaml:", err)
		}
	case ".toml":
		if err = toml.Unmarshal(data, &cfg); err != nil {
			log.Fatal("failed to parse toml: ", err)
		}
	default:
		log.Fatal("unknown config file extension:", filepath.Ext(path))
	}

	// Loading core-level plugins.
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

		registerActiveCorePlugin(pcfg.Name, p)
	}

	return cfg
}
