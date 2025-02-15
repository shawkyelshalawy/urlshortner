package app

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shawkyelshalawy/urlshortner/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	Redis         *redis.Client
	Mongo         *mongo.Client
	AnalyticsChan chan AnalyticsEvent
}

type AnalyticsEvent struct {
	ShortID    string
	ClientIP   string
	UserAgent  string
	Timestamp  time.Time
}