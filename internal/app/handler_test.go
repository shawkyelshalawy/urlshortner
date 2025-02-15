package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/shawkyelshalawy/urlshortner/internal/app"
	"github.com/shawkyelshalawy/urlshortner/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	ShortURL string `json:"short_url"`
	Error    string `json:"error,omitempty"`
}

func setupTest(t *testing.T) (*app.Application, func()) {
	
	gin.SetMode(gin.TestMode)

	
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, //separate DB for testing
	})

	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, rdb.Ping(ctx).Err(), "Redis connection failed")

	app := &app.Application{
		Config: &config.Config{
			BaseURL: "http://localhost:8080",
		},
		Redis:         rdb,
		AnalyticsChan: make(chan app.AnalyticsEvent, 1),
	}

	// Return cleanup function
	cleanup := func() {
		rdb.FlushDB(context.Background())
		rdb.Close()
		close(app.AnalyticsChan)
	}

	return app, cleanup
}

func TestShortenHandler(t *testing.T) {
	app, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		setupFunc      func(context.Context) error
		validateFunc   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Create new short URL",
			payload:        `{"url":"https://example.com"}`,
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp testResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Contains(t, resp.ShortURL, "http://localhost:8080/")
			},
		},
		{
			name:           "Return existing short URL",
			payload:        `{"url":"https://existing.com"}`,
			expectedStatus: http.StatusOK,
			setupFunc: func(ctx context.Context) error {
				return app.Redis.Set(ctx, "long:https://existing.com", "abc123", 0).Err()
			},
			validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp testResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)
			},
		},
		{
			name:           "Reject invalid URL",
			payload:        `{"url":"not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid request payload")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				require.NoError(t, tt.setupFunc(ctx))
				cancel()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.payload))
			c.Request.Header.Set("Content-Type", "application/json")

			app.ShortenHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, w)
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	app, cleanup := setupTest(t)
	defer cleanup()

	ctx := context.Background()
	shortID := "test123"
	originalURL := "https://original.com"

	// Seed test data
	require.NoError(t, app.Redis.Set(ctx, "short:"+shortID, originalURL, 0).Err())

	tests := []struct {
		name           string
		shortID        string
		expectedStatus int
		expectAnalytic bool
		expectedURL    string
	}{
		{
			name:           "Valid redirect",
			shortID:        shortID,
			expectedStatus: http.StatusMovedPermanently,
			expectAnalytic: true,
			expectedURL:    originalURL,
		},
		{
			name:           "Not found",
			shortID:        "invalid",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.shortID, nil)
			c.Params = []gin.Param{{Key: "shortID", Value: tt.shortID}}

			app.RedirectHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedURL != "" {
				assert.Equal(t, tt.expectedURL, w.Header().Get("Location"))
			}

			if tt.expectAnalytic {
				select {
				case event := <-app.AnalyticsChan:
					assert.Equal(t, tt.shortID, event.ShortID)
					assert.NotEmpty(t, event.ClientIP)
				case <-time.After(time.Second):
					t.Error("Expected analytics event but none received")
				}
			}
		})
	}
}