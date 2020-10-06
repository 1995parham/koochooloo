package metric

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1995parham/koochooloo/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server contains information about metrics server.
type Server struct {
	Address string
	Enabled bool
}

// NewServer creates a new monitoring server.
func NewServer(cfg config.Monitoring) Server {
	return Server{
		Address: cfg.Address,
		Enabled: cfg.Enabled,
	}
}

// Start creates and run a metric server for prometheus in new go routine.
func (s Server) Start() {
	if s.Enabled {
		srv := http.NewServeMux()
		srv.Handle("/metrics", promhttp.Handler())

		go func() {
			if err := http.ListenAndServe(s.Address, srv); !errors.Is(err, http.ErrServerClosed) {
				fmt.Println(err)
			}
		}()
	}
}
