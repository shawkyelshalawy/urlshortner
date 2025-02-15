package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func (app *Application) rateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !app.Config.RateLimit.Enabled {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := "rate_limit:" + ip

		current, err := app.Redis.Get(c, key).Int()
		if err != nil && err != redis.Nil {
			app.Logger.Error("Rate limiter error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if current >= app.Config.RateLimit.Requests {
			retryAfter := app.Config.RateLimit.Window.Seconds()
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		pipe := app.Redis.TxPipeline()
		pipe.Incr(c, key)
		pipe.Expire(c, key, app.Config.RateLimit.Window)
		if _, err := pipe.Exec(c); err != nil {
			app.Logger.Error("Rate limiter error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
	}
}