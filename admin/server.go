package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/starwalkn/kairyu"
)

func StartAdminServer(cfg *kairyu.GatewayConfig, port int) {
	mux := http.NewServeMux()

	// Отдаём конфиг как JSON
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	})

	// Отдаём статические файлы (index.html, css, js)
	staticDir := filepath.Join(".", "admin", "static")
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("📊 Admin dashboard available at http://localhost%s\n", addr)
	go http.ListenAndServe(addr, mux)
}
