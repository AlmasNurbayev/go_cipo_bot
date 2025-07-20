package botP

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
)

type HttpApp struct {
	Log    *slog.Logger
	Server *http.Server
	Cfg    *config.Config
}

func NewHttpApp(cfg *config.Config, log *slog.Logger) (*HttpApp, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTP_PORT,
		Handler: mux,
	}
	return &HttpApp{
		Log:    log,
		Server: server,
		Cfg:    cfg,
	}, nil
}

func (h *HttpApp) Run() {
	h.Log.Info("HTTP server started", slog.String("port", h.Cfg.HTTP_PORT))
	if err := h.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.Log.Error("listen error", slog.Any("err", err))
	}
}

func (h *HttpApp) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), h.Cfg.BOT_TIMEOUT)
	defer cancel()
	if err := h.Server.Shutdown(ctx); err != nil {
		h.Log.Error("HTTP server shutdown error", slog.Any("err", err))
	}
}
