package imparrot

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

// ClientPool manages a pool of IM clients with automatic resource management
// This is especially useful for message forwarding servers that need to handle
// multiple bots and prevent resource leaks
type ClientPool struct {
	clients   map[string]types.IMParrot
	mu        sync.RWMutex
	maxIdle   time.Duration
	lastUsed  map[string]time.Time
	httpPool  *http.Client // Shared HTTP client for all connections
	closeChan chan struct{}
	wg        sync.WaitGroup
}

// PoolConfig configures the client pool
type PoolConfig struct {
	// MaxIdleTime is the maximum time a client can remain idle before being closed
	// Default: 30 minutes
	MaxIdleTime time.Duration

	// CleanupInterval is how often to check for idle clients
	// Default: 5 minutes
	CleanupInterval time.Duration

	// HTTPClientConfig for the shared HTTP client
	HTTPTimeout time.Duration
	// MaxIdleConns controls the maximum number of idle (keep-alive) connections
	MaxIdleConns int
	// MaxIdleConnsPerHost controls the maximum idle connections per host
	MaxIdleConnsPerHost int
}

// DefaultPoolConfig returns a pool config with sensible defaults
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxIdleTime:         30 * time.Minute,
		CleanupInterval:     5 * time.Minute,
		HTTPTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}
}

// NewClientPool creates a new client pool with automatic resource management
func NewClientPool(config *PoolConfig) *ClientPool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	// Create shared HTTP client with proper connection pooling
	httpClient := &http.Client{
		Timeout: config.HTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	pool := &ClientPool{
		clients:   make(map[string]types.IMParrot),
		lastUsed:  make(map[string]time.Time),
		maxIdle:   config.MaxIdleTime,
		httpPool:  httpClient,
		closeChan: make(chan struct{}),
	}

	// Start background cleanup goroutine
	pool.wg.Add(1)
	go pool.cleanupLoop(config.CleanupInterval)

	return pool
}

// GetOrCreate gets an existing client or creates a new one
// The key is used to identify the client (e.g., "platform:appid" or "bottoken")
func (p *ClientPool) GetOrCreate(ctx context.Context, key string, platform string, config types.Config) (types.IMParrot, error) {
	// Try to get existing client
	p.mu.RLock()
	if client, ok := p.clients[key]; ok {
		p.mu.RUnlock()
		// Update last used time
		p.mu.Lock()
		p.lastUsed[key] = time.Now()
		p.mu.Unlock()
		return client, nil
	}
	p.mu.RUnlock()

	// Create new client
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check in case another goroutine created it
	if client, ok := p.clients[key]; ok {
		p.lastUsed[key] = time.Now()
		return client, nil
	}

	// Create new client with shared HTTP client
	client, err := p.createClient(platform, config)
	if err != nil {
		return nil, err
	}

	p.clients[key] = client
	p.lastUsed[key] = time.Now()

	return client, nil
}

// Get retrieves a client by key without creating it
func (p *ClientPool) Get(key string) (types.IMParrot, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	client, ok := p.clients[key]
	if ok {
		p.mu.RUnlock()
		p.mu.Lock()
		p.lastUsed[key] = time.Now()
		p.mu.Unlock()
		p.mu.RLock()
	}
	return client, ok
}

// Remove removes and closes a client
func (p *ClientPool) Remove(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, ok := p.clients[key]
	if !ok {
		return nil
	}

	delete(p.clients, key)
	delete(p.lastUsed, key)

	return client.Close()
}

// Size returns the current number of clients in the pool
func (p *ClientPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

// cleanupLoop periodically removes idle clients
func (p *ClientPool) cleanupLoop(interval time.Duration) {
	defer p.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanupIdle()
		case <-p.closeChan:
			return
		}
	}
}

// cleanupIdle removes clients that have been idle for too long
func (p *ClientPool) cleanupIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	toRemove := []string{}

	for key, lastUsed := range p.lastUsed {
		if now.Sub(lastUsed) > p.maxIdle {
			toRemove = append(toRemove, key)
		}
	}

	for _, key := range toRemove {
		if client, ok := p.clients[key]; ok {
			_ = client.Close() // Ignore error during cleanup
			delete(p.clients, key)
			delete(p.lastUsed, key)
		}
	}

	if len(toRemove) > 0 {
		fmt.Printf("ClientPool: cleaned up %d idle clients\n", len(toRemove))
	}
}

// Close closes all clients and stops the cleanup loop
func (p *ClientPool) Close() error {
	// Stop cleanup loop
	close(p.closeChan)
	p.wg.Wait()

	// Close all clients
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for key, client := range p.clients {
		if err := client.Close(); err != nil {
			lastErr = err
		}
		delete(p.clients, key)
		delete(p.lastUsed, key)
	}

	// Close shared HTTP client connections
	p.httpPool.CloseIdleConnections()

	return lastErr
}

// createClient creates a client with shared HTTP client
func (p *ClientPool) createClient(platform string, config types.Config) (types.IMParrot, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if config.GetPlatform() != platform {
		return nil, fmt.Errorf("config platform %s does not match requested platform %s",
			config.GetPlatform(), platform)
	}

	// Use shared HTTP client for all platforms
	return createClientWithHTTP(platform, config, p.httpPool)
}
