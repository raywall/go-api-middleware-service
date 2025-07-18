// Package middleware provides tests for the middleware functions of the pricing microservice.
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGuardrails tests the Guardrails middleware for various scenarios.
func TestGuardrails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ValidRate", func(t *testing.T) {
		var capturedContext *gin.Context
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("pricing_result", 5.0)
			c.Next()
		})
		router.Use(Guardrails())
		router.GET("/test", func(c *gin.Context) {
			capturedContext = c
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200, got %d: %s", w.Code, w.Body.String())
		assert.NotNil(t, capturedContext, "Expected captured context")
		validatedRate, exists := capturedContext.Get("validated_rate")
		assert.True(t, exists, "Expected validated_rate in context")
		assert.Equal(t, 5.0, validatedRate, "Expected validated_rate 5.0, got %v", validatedRate)
	})

	t.Run("MissingPricingResult", func(t *testing.T) {
		router := gin.New()
		router.Use(Guardrails())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status 500, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "no pricing result")
	})

	t.Run("InvalidRate", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("pricing_result", 25.0)
			c.Next()
		})
		router.Use(Guardrails())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status 500, got %d: %s", w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "invalid rate")
	})
}
