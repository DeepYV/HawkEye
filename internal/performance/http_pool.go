/**
 * HTTP Client Pool
 * 
 * Author: Principal Engineer + Team Beta
 * Responsibility: HTTP client connection pooling and reuse
 */

package performance

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// HTTPClientPool manages a pool of HTTP clients with connection reuse
type HTTPClientPool struct {
	clients map[string]*http.Client
	mu      sync.RWMutex
}

// NewHTTPClientPool creates a new HTTP client pool
func NewHTTPClientPool() *HTTPClientPool {
	return &HTTPClientPool{
		clients: make(map[string]*http.Client),
	}
}

// GetClient gets or creates an HTTP client for a base URL
func (p *HTTPClientPool) GetClient(baseURL string) *http.Client {
	p.mu.RLock()
	if client, exists := p.clients[baseURL]; exists {
		p.mu.RUnlock()
		return client
	}
	p.mu.RUnlock()

	// Create new client with connection pooling
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := p.clients[baseURL]; exists {
		return client
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		ResponseHeaderTimeout:  10 * time.Second,
		DisableKeepAlives:      false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	p.clients[baseURL] = client
	return client
}

// Close closes all clients in the pool
func (p *HTTPClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, client := range p.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
