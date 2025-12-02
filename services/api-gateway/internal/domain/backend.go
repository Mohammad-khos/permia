package domain

// BackendService defines a backend microservice that the gateway routes to
type BackendService struct {
	Name     string
	URL      string
	BasePath string
	Timeout  int // in seconds
}

// GatewayRequest represents a request coming to the gateway
type GatewayRequest struct {
	Method  string
	Path    string
	Query   string
	Headers map[string][]string
	Body    []byte
}

// GatewayResponse represents a response from the gateway
type GatewayResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}
