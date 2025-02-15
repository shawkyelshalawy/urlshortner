package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shawkyelshalawy/urlshortner/internal/config"
	"github.com/shawkyelshalawy/urlshortner/internal/handlers"
	"github.com/shawkyelshalawy/urlshortner/internal/storage"
)

func main() {
	cfg := config.Load()

	rdb , err := storage.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/shorten", func(c *gin.Context) {
		handlers.ShortenHandler(c, rdb, cfg.BaseURL)
	})

	router.GET("/:shortID", func(c *gin.Context) {
		handlers.RedirectHandler(c, rdb)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}