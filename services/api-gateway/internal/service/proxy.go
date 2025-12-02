package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// ProxyService handles routing requests to backend services
type ProxyService struct {
	client         *resty.Client
	logger         *zap.SugaredLogger
	coreServiceURL string
	botServiceURL  string
}

// NewProxyService creates a new ProxyService
func NewProxyService(coreURL, botURL string, logger *zap.SugaredLogger) *ProxyService {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second)

	return &ProxyService{
		client:         client,
		logger:         logger,
		coreServiceURL: coreURL,
		botServiceURL:  botURL,
	}
}

// RouteToCore routes a request to the core service
func (p *ProxyService) RouteToCore(method, path string, headers http.Header, body []byte) (*http.Response, error) {
	return p.route(p.coreServiceURL, method, path, headers, body)
}

// RouteToBot routes a request to the bot service
func (p *ProxyService) RouteToBot(method, path string, headers http.Header, body []byte) (*http.Response, error) {
	return p.route(p.botServiceURL, method, path, headers, body)
}

// route is the generic routing function
func (p *ProxyService) route(baseURL, method, path string, headers http.Header, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", baseURL, path)

	p.logger.Infof("Routing %s %s to %s", method, path, baseURL)

	req := p.client.R()

	// Copy relevant headers
	for key, values := range headers {
		if len(values) > 0 {
			req.SetHeader(key, values[0])
		}
	}

	// Set body if present
	if body != nil {
		req.SetBody(body)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		p.logger.Errorf("Routing error to %s: %v", baseURL, err)
		return nil, err
	}

	p.logger.Debugf("Response from %s: %d", baseURL, resp.StatusCode())
	return resp.RawResponse, nil
}

// HealthCheck checks if backend services are healthy
func (p *ProxyService) HealthCheck(serviceURL string) bool {
	resp, err := p.client.R().Get(fmt.Sprintf("%s/health", serviceURL))
	if err != nil {
		p.logger.Warnf("Health check failed for %s: %v", serviceURL, err)
		return false
	}
	return resp.StatusCode() == http.StatusOK
}
