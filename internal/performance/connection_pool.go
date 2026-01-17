/**
 * Connection Pooling
 * 
 * Author: Principal Engineer + Team Beta
 * Responsibility: Database connection pooling for performance
 */

package performance

import (
	"context"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/your-org/frustration-engine/internal/config"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ConnectionPool manages a pool of ClickHouse connections
type ConnectionPool struct {
	dsn          string
	poolSize     int
	connections  chan driver.Conn
	mu           sync.RWMutex
	activeCount  int
	maxRetries   int
	retryBackoff time.Duration
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(dsn string, poolSize int) (*ConnectionPool, error) {
	if poolSize <= 0 {
		poolSize = 10 // Default pool size
	}

	pool := &ConnectionPool{
		dsn:          dsn,
		poolSize:     poolSize,
		connections:  make(chan driver.Conn, poolSize),
		maxRetries:   3,
		retryBackoff: 100 * time.Millisecond,
	}

	// Pre-populate pool
	for i := 0; i < poolSize; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			return nil, err
		}
		pool.connections <- conn
		pool.activeCount++
	}

	return pool, nil
}

// GetConnection gets a connection from the pool
func (p *ConnectionPool) GetConnection(ctx context.Context) (driver.Conn, error) {
	select {
	case conn := <-p.connections:
		// Check if connection is still alive
		if err := conn.Ping(ctx); err != nil {
			// Connection is dead, create new one
			return p.createConnection()
		}
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Pool exhausted, create new connection
		return p.createConnection()
	}
}

// ReturnConnection returns a connection to the pool
func (p *ConnectionPool) ReturnConnection(conn driver.Conn) {
	select {
	case p.connections <- conn:
		// Returned to pool
	default:
		// Pool full, close connection
		conn.Close()
		p.mu.Lock()
		p.activeCount--
		p.mu.Unlock()
	}
}

// createConnection creates a new ClickHouse connection
func (p *ConnectionPool) createConnection() (driver.Conn, error) {
	var conn driver.Conn
	var err error

	// Parse DSN - handle both "clickhouse://host:port" and "host:port" formats
	addr := p.dsn
	if len(p.dsn) > 12 && p.dsn[:12] == "clickhouse://" {
		// Remove "clickhouse://" prefix
		addr = p.dsn[12:]
	}

	// Get password from environment if set, otherwise use empty (default ClickHouse setup)
	password := config.GetEnv("CLICKHOUSE_PASSWORD", "")
	username := config.GetEnv("CLICKHOUSE_USERNAME", "default")

	for i := 0; i < p.maxRetries; i++ {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{addr},
			Auth: clickhouse.Auth{
				Database: "events",
				Username: username,
				Password: password,
			},
			Settings: clickhouse.Settings{
				"max_execution_time": 60,
			},
			DialTimeout: 5 * time.Second,
		})

		if err == nil {
			p.mu.Lock()
			p.activeCount++
			p.mu.Unlock()
			return conn, nil
		}

		// Exponential backoff
		time.Sleep(p.retryBackoff * time.Duration(1<<uint(i)))
	}

	return nil, err
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
	return nil
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() (active, available int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.activeCount, len(p.connections)
}
