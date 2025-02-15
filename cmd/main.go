package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shawkyelshalawy/urlshortner/internal/app"
	"github.com/shawkyelshalawy/urlshortner/internal/config"
	"github.com/shawkyelshalawy/urlshortner/internal/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	
	rdb, err := storage.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	
	mongoClient, err := storage.NewMongoClient(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	application := &app.Application{
		Config:        cfg,
		Logger:        logger,
		Redis:         rdb,
		Mongo:         mongoClient,
		AnalyticsChan: make(chan app.AnalyticsEvent, 100),
	}

	
	go application.StartAnalyticsWorker(context.Background())

	
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      application.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("Shutting down server", zap.String("signal", s.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	logger.Info("Starting server", zap.String("addr", srv.Addr))

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Server failed", zap.Error(err))
	}

	err = <-shutdownError
	if err != nil {
		logger.Fatal("Graceful shutdown failed", zap.Error(err))
	}

	logger.Info("Server stopped")
}