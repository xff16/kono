package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/starwalkn/kairyu"
	"github.com/starwalkn/kairyu/admin"
)

func main() {
	data, err := os.ReadFile("kairyu.json")
	if err != nil {
		log.Fatal("failed to read config:", err)
	}

	var cfg kairyu.GatewayConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("failed to parse config:", err)
	}

	router := kairyu.NewRouter(cfg.Routes)

	addr := ":8080"
	fmt.Println("🚀 Kairyu API Gateway started at", addr)
	PrintConfig(cfg)

	go admin.StartAdminServer(&cfg, cfg.Server.Port+1)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

var dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Kairyu Dashboard</title>
	<style>
		body {
			font-family: Inter, sans-serif;
			background: #0d1117;
			color: #c9d1d9;
			margin: 0;
			padding: 0;
		}
		header {
			background: #161b22;
			padding: 1rem 2rem;
			font-size: 1.2rem;
			border-bottom: 1px solid #30363d;
		}
		main {
			padding: 2rem;
		}
		pre {
			background: #161b22;
			border-radius: 10px;
			padding: 1rem;
			overflow-x: auto;
			color: #58a6ff;
		}
	</style>
</head>
<body>
	<header>
		Kairyu 🐉 Dashboard
	</header>
	<main>
		<h2>Current Configuration</h2>
		<pre id="config">Loading...</pre>
	</main>
	<script>
		fetch('/api/config')
			.then(res => res.json())
			.then(cfg => {
				document.getElementById('config').textContent = JSON.stringify(cfg, null, 2);
			})
			.catch(err => {
				document.getElementById('config').textContent = 'Failed to load configuration: ' + err;
			});
	</script>
</body>
</html>
`

func PrintConfig(cfg kairyu.GatewayConfig) {
	fmt.Printf("\n🚀 Loaded Kairyu Gateway configuration:\n")
	fmt.Printf("────────────────────────────────────────────\n")
	fmt.Printf("Name:    %s\n", cfg.Name)
	fmt.Printf("Version: %s\n", cfg.Version)
	fmt.Printf("Schema:  %s\n", cfg.Schema)
	fmt.Printf("Server:  port=%d  timeout=%dms\n\n", cfg.Server.Port, cfg.Server.Timeout)

	if len(cfg.Plugins) > 0 {
		fmt.Println("🔌 Global plugins:")
		for _, p := range cfg.Plugins {
			fmt.Printf("   • %s\n", p.Name)
		}
		fmt.Println()
	}

	fmt.Println("📦 Routes:")
	for i, route := range cfg.Routes {
		fmt.Printf(" %d) %s %s\n", i+1, route.Method, route.Path)

		if len(route.Plugins) > 0 {
			fmt.Println("    ├─ Plugins:")
			for _, p := range route.Plugins {
				fmt.Printf("    │   • %s\n", p.Name)
				if len(p.Config) > 0 {
					fmt.Printf("    │     Config:\n")
					for k, v := range p.Config {
						fmt.Printf("    │       - %s: %v\n", k, v)
					}
				}
			}
		}

		if len(route.Backends) > 0 {
			fmt.Println("    ├─ Backends:")
			for _, b := range route.Backends {
				timeout := ""
				if b.Timeout > 0 {
					timeout = fmt.Sprintf(" (timeout=%dms)", b.Timeout)
				}
				fmt.Printf("    │   • %s %s%s\n", b.Method, b.URL, timeout)
			}
		}

		if route.Aggregate != "" || route.Transform != "" {
			fmt.Println("    └─ Options:")
			if route.Aggregate != "" {
				fmt.Printf("        • aggregate: %s\n", route.Aggregate)
			}
			if route.Transform != "" {
				fmt.Printf("        • transform: %s\n", route.Transform)
			}
		}
		fmt.Println()
	}
	fmt.Printf("────────────────────────────────────────────\n\n")
}
