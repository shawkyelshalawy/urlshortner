package app

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaevor/go-nanoid"
	"go.uber.org/zap"
)

func (app *Application) ShortenHandler(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		app.errorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	
	existing, err := app.Redis.HGet(c, os.Getenv("shortenedUrlsHash"), req.URL).Result()
	if err == nil && existing != "" {
		c.JSON(http.StatusOK, gin.H{"short_url": app.Config.BaseURL + "/" + existing})
		return
	}

	
	gen, _ := nanoid.Standard(6)
	shortID := gen()

	pipe := app.Redis.TxPipeline()
	pipe.HSet(c,  os.Getenv("shortenedUrlsHash"), shortID, req.URL)
	pipe.HSet(c, os.Getenv("originalUrlsHash"), req.URL, shortID)
	
	pipe.Expire(c, os.Getenv("shortenedUrlsHash"), 30*24*time.Hour)
	pipe.Expire(c, os.Getenv("originalUrlsHash"), 30*24*time.Hour)

	if _, err := pipe.Exec(c); err != nil {
		app.errorResponse(c, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_url":  app.Config.BaseURL + "/" + shortID,
		"expires_at": time.Now().Add(30 * 24 * time.Hour),
	})
}

func (app *Application) RedirectHandler(c *gin.Context) {
	shortID := c.Param("shortID")
	
	app.Logger.Info("Redirect request received",
		zap.String("shortID", shortID),
		zap.String("path", c.Request.URL.Path),
	)

	
	url, err := app.Redis.HGet(c, os.Getenv("shortenedUrlsHash"), shortID).Result()
	if err != nil {
		app.Logger.Error("Redis get error",
			zap.String("shortID", shortID),
			zap.Error(err),
		)
		app.errorResponse(c, http.StatusNotFound, "URL not found")
		return
	}

	app.AnalyticsChan <- AnalyticsEvent{
		ShortID:    shortID,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Timestamp:  time.Now(),
	}

	c.Redirect(http.StatusMovedPermanently, url)
}

func (app *Application) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}