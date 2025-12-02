package handler

import (
	"io"
	"net/http"

	"Permia/api-gateway/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests
type Handler struct {
	proxyService *service.ProxyService
	logger       *zap.SugaredLogger
}

// NewHandler creates a new Handler
func NewHandler(proxyService *service.ProxyService, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// Health returns the health status of the gateway
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "Permia API Gateway",
	})
}

// ProxyCoreAPI proxies requests to the core service
func (h *Handler) ProxyCoreAPI(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Errorf("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Route to core service
	resp, err := h.proxyService.RouteToCore(method, path, c.Request.Header, body)
	if err != nil {
		h.logger.Errorf("Failed to proxy to core service: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Core service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Copy headers from upstream response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ProxyBotAPI proxies requests to the bot service
func (h *Handler) ProxyBotAPI(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Errorf("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Route to bot service
	resp, err := h.proxyService.RouteToBot(method, path, c.Request.Header, body)
	if err != nil {
		h.logger.Errorf("Failed to proxy to bot service: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Bot service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Errorf("Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Copy headers from upstream response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// NotFound handles 404 requests
func (h *Handler) NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Endpoint not found",
		"path":  c.Request.URL.Path,
	})
}
