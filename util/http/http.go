package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-manager/util/errs"
)

func RunHttpServer(httpServer *http.Server) error {
	slog.Info("Starting http server", "addr", httpServer.Addr)
	httpServerError := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("Http server closed")
			} else {
				httpServerError <- errs.Wrap("httpServer stopped with unexpected error", err)
			}
		}
	}()

	// Block until shutdown signal or server error is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-signals:
		slog.Info("Shutdown signal received. About to shut down")

		slog.Info("Shutting down http server down gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			return errs.Wrap("httpServer shutdown error", err)
		}

		slog.Info("Http server has shutdown normally")
	case err := <-httpServerError:
		return err
	}

	return nil
}
