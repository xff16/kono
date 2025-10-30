package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/starwalkn/bravka"
	"github.com/starwalkn/bravka/dashboard"

	_ "github.com/starwalkn/bravka/internal/plugin/ratelimit"
)

func main() {
	cfgPath := os.Getenv("BRAVKA_CONFIG")
	if cfgPath == "" {
		cfgPath = "./bravka.json"
	}

	cfg := bravka.LoadConfig(cfgPath)

	if cfg.Dashboard.Enable {
		adminServer := dashboard.NewServer(&cfg)
		go adminServer.Start()
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      bravka.NewRouter(cfg.Routes),
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
