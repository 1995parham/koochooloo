package metric

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server contains information about metrics server.
type Server struct {
	srv     *http.ServeMux
	address string
}

// NewServer creates a new monitoring server.
func NewServer(cfg *Config) Server {
	var srv *http.ServeMux

	if cfg.Enabled {
		srv = http.NewServeMux()
		srv.Handle("/metrics", promhttp.Handler())
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return Server{address: addr, srv: srv}
}

// Serve creates and run a metric server for prometheus.
func (s Server) Serve() error {
	if s.srv == nil {
		return nil
	}
	return http.ListenAndServe(s.address, s.srv)
}
