package admin

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/starwalkn/bravka"
	"github.com/starwalkn/bravka/internal/logger"
)

type Server struct {
	cfg *bravka.GatewayConfig
	log *zap.Logger
}

func NewServer(cfg *bravka.GatewayConfig) *Server {
	return &Server{
		cfg: cfg,
		log: logger.Init(true).Named("admin-panel"),
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	addr := fmt.Sprintf(":%d", s.cfg.Dashboard.Port)

	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	s.log.Info("📊 Admin dashboard started\n", zap.String("addr", addr))

	if err := server.ListenAndServe(); err != nil {
		s.log.Error("admin server had errors, disabling", zap.Error(err))
		return
	}
}
