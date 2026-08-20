package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/database"
	"starter-backend/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	db, err := database.Open(cfg.DatabaseDSN, cfg.AppEnv)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)
	if cfg.AutoMigrate {
		if err := database.Migrate(db); err != nil {
			logger.Error("database migration failed", "error", err)
			os.Exit(1)
		}
	}

	handler, err := httpapi.New(db, cfg, logger)
	if err != nil {
		fail(err)
	}
	server := &http.Server{
		Addr: cfg.ServerHost + ":" + cfg.ServerPort, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownSignal.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}()

	logger.Info("server started", "address", server.Addr, "environment", cfg.AppEnv)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func fail(err error) {
	slog.Error("startup failed", "error", err)
	os.Exit(1)
}
