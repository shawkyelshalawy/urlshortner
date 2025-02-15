package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jaevor/go-nanoid"
)

const urlTTL = 30 * 24 * time.Hour

type shortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func ShortenHandler(c *gin.Context, rdb *redis.Client, baseURL string) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrInvalidPayload)
		return
	}

	ctx := context.Background()

	existing, err := rdb.Get(ctx, "long:"+req.URL).Result()
	if err == nil && existing != "" {
		c.JSON(http.StatusOK, shortenResponse{ShortURL: baseURL + existing})
		return
	}

	gen, err := nanoid.Standard(6)
	if err != nil {
		serverErrorResponse(c, err)
		return
	}
	shortID := gen()

	if err := rdb.Set(ctx, "short:"+shortID, req.URL, urlTTL).Err(); err != nil {
		serverErrorResponse(c, err)
		return
	}
	if err := rdb.Set(ctx, "long:"+req.URL, shortID, urlTTL).Err(); err != nil {
		rdb.Del(ctx, "short:"+shortID)
		serverErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, shortenResponse{ShortURL: baseURL + shortID})
}

func RedirectHandler(c *gin.Context, rdb *redis.Client) {
	shortID := c.Param("shortID")
	ctx := context.Background()

	longURL, err := rdb.Get(ctx, "short:"+shortID).Result()
	if err != nil {
		notFoundResponse(c)
		return
	}

	c.Redirect(http.StatusMovedPermanently, longURL)
}
