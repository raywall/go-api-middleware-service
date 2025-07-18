// Package middleware provides tests for the middleware functions of the pricing microservice.
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockStatsdClient is a mock for statsd.Client to capture metrics.
type mockStatsdClient struct {
	counters   map[string]int
	histograms map[string][]float64
}

func (m *mockStatsdClient) Incr(name string, tags []string, rate float64) error {
	m.counters[name]++
	return nil
}

func (m *mockStatsdClient) Histogram(name string, value float64, tags []string, rate float64) error {
	m.histograms[name] = append(m.histograms[name], value)
	return nil
}

// TestObservability tests the Observability middleware for metric logging.
func TestObservability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		// Initialize mock Datadog client
		client := &mockStatsdClient{
			counters:   make(map[string]int),
			histograms: make(map[string][]float64),
		}

		// Set up router
		router := gin.New()
		router.Use(Observability(client))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		// Assert metrics
		assert.Equal(t, 1, client.counters["request.count"])
		assert.Len(t, client.histograms["request.duration"], 1)
		assert.Equal(t, 0, client.counters["request.error"])
	})

	t.Run("Error", func(t *testing.T) {
		client := &mockStatsdClient{
			counters:   make(map[string]int),
			histograms: make(map[string][]float64),
		}

		router := gin.New()
		router.Use(Observability(client))
		router.GET("/test", func(c *gin.Context) {
			c.Error(errors.New("test error"))
			c.AbortWithStatus(http.StatusInternalServerError)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, 1, client.counters["request.count"])
		assert.Len(t, client.histograms["request.duration"], 1)
		assert.Equal(t, 1, client.counters["request.error"])
	})
}
