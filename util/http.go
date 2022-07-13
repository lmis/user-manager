package util

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunHttpServer(log Logger, httpServer *http.Server) error {
	log.Info("Starting http server on %s", httpServer.Addr)
	httpServerError := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("Http server closed")
			} else {
				httpServerError <- Wrap("httpServer stopped with unexpected error", err)
			}
		}
	}()

	// Block until shutdown signal or server error is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-signals:
		log.Info("Shutdown signal received. About to shut down")

		log.Info("Shutting down http server down gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			return Wrap("httpServer shutdown error", err)
		}

		log.Info("Http server has shutdown normally")
	case err := <-httpServerError:
		return err
	}

	return nil
}
