package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaevor/go-nanoid"
)

func (app *Application) shortenHandler(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		app.errorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	
	existing, err := app.Redis.Get(c, "long:"+req.URL).Result()
	if err == nil && existing != "" {
		c.JSON(http.StatusOK, gin.H{"short_url": app.Config.BaseURL + "/" + existing})
		return
	}

	
	gen, _ := nanoid.Standard(6)
	shortID := gen()

	
	pipe := app.Redis.TxPipeline()
	pipe.Set(c, "short:"+shortID, req.URL, 30*24*time.Hour)
	pipe.Set(c, "long:"+req.URL, shortID, 30*24*time.Hour)
	if _, err := pipe.Exec(c); err != nil {
		app.errorResponse(c, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_url": app.Config.BaseURL + "/" + shortID,
		"expires_at": time.Now().Add(30 * 24 * time.Hour),
	})
}

func (app *Application) redirectHandler(c *gin.Context) {
	shortID := c.Param("shortID")
	url, err := app.Redis.Get(c, "short:"+shortID).Result()
	if err != nil {
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