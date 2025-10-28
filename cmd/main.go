package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/starwalkn/bravka"
	"github.com/starwalkn/bravka/admin"

	_ "github.com/starwalkn/bravka/internal/plugin/ratelimit"
)

func main() {
	cfgPath := os.Getenv("BRAVKA_CONFIG")
	cfg := bravka.LoadConfig(cfgPath)

	if cfg.Dashboard.Enable {
		adminServer := admin.NewServer(&cfg)
		go adminServer.Start()
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      bravka.NewRouter(cfg.Routes),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
