package kairyu

import (
	"encoding/json"
	"log"
	"os"
)

type GatewayConfig struct {
	Schema  string         `json:"schema"`
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Server  ServerConfig   `json:"server"`
	Plugins []PluginConfig `json:"plugins"`
	Routes  []RouteConfig  `json:"routes"`
}

type ServerConfig struct {
	Port    int `json:"port"`
	Timeout int `json:"timeout"`
}

type RouteConfig struct {
	Path      string          `json:"path"`
	Method    string          `json:"method"`
	Plugins   []PluginConfig  `json:"plugins"`
	Backends  []BackendConfig `json:"backends"`
	Aggregate string          `json:"aggregate"`
	Transform string          `json:"transform"`
}

type BackendConfig struct {
	URL     string `json:"url"`
	Method  string `json:"method"`
	Timeout int    `json:"timeout"`
}

type PluginConfig struct {
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

func LoadConfig(path string) GatewayConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read config file:", err)
	}

	var cfg GatewayConfig
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("Failed to parse YAML:", err)
	}

	return cfg
}
