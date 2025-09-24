package httpserver

import (
	"context"
	"net"
	"net/http"
	"task-processor/internal/infrastructure/config"
)

type HTTPServer struct {
	server *http.Server
	cfg    *config.Config
}

func NewHTTPServer(
	cfg 	*config.Config, 
	handler  http.Handler, 
	baseCtx  context.Context, 
) *HTTPServer {
	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		BaseContext:  func(_ net.Listener) context.Context { return baseCtx },
	}
	return &HTTPServer{
		server: srv, 
		cfg: cfg, 
	}	
}

func (h *HTTPServer) ListenAndServe() error {
	return h.server.ListenAndServe()
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTPServer) Addr() string {
	return h.server.Addr
}
