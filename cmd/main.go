package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/starwalkn/tokka"
	"github.com/starwalkn/tokka/dashboard"
	"github.com/starwalkn/tokka/internal/logger"
	_ "github.com/starwalkn/tokka/internal/plugin/ratelimit"
)

func main() {
	cfgPath := os.Getenv("TOKKA_CONFIG")
	if cfgPath == "" {
		cfgPath = "./tokka.json"
	}

	cfg := tokka.LoadConfig(cfgPath)

	log := logger.New(cfg.Debug)
	defer func() {
		if err := log.Sync(); err != nil {
			log.Warn("cannot sync log", zap.Error(err))
		}
	}()

	if cfg.Dashboard.Enable {
		dashboardServer := dashboard.NewServer(&cfg, log.Named("dashboard"))
		go dashboardServer.Start()
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      tokka.NewRouter(cfg.Routes, cfg.Middlewares, log.Named("router")),
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
