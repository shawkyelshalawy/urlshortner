package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *Application) Routes() *gin.Engine {
	router := gin.New()

	
	router.Use(app.rateLimiter())
	router.Use(gin.Recovery())

	// Routes
	router.GET("/ping", app.pingHandler)
	router.POST("/shorten", app.shortenHandler)
	router.GET("/:shortID", app.redirectHandler)

	return router
}

func (app *Application) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}