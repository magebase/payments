package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ConnectionManager manages database connections
type ConnectionManager struct {
	yugabytePool *pgxpool.Pool
	clickHouse   clickhouse.Conn
	tracer       trace.Tracer
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		tracer: otel.Tracer("payments.db"),
	}
}

// ConnectYugabyte connects to Yugabyte DB
func (cm *ConnectionManager) ConnectYugabyte(ctx context.Context, config *YugabyteConfig) error {
	ctx, span := cm.tracer.Start(ctx, "ConnectYugabyte")
	defer span.End()

	log.Printf("Connecting to Yugabyte DB at %s:%d", config.Host, config.Port)

	poolConfig, err := pgxpool.ParseConfig(config.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(config.MaxConnections)
	poolConfig.MinConns = int32(config.MinConnections)
	poolConfig.MaxConnIdleTime = time.Duration(config.MaxIdleTime) * time.Second

	// Add connection hooks for tracing
	poolConfig.BeforeConnect = func(ctx context.Context, cc *pgx.ConnConfig) error {
		log.Printf("Connecting to database: %s", cc.Database)
		return nil
	}

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		log.Printf("Successfully connected to database: %s", conn.Config().Database)
		return nil
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	cm.yugabytePool = pool
	log.Printf("Successfully connected to Yugabyte DB")
	return nil
}

// ConnectClickHouse connects to ClickHouse
func (cm *ConnectionManager) ConnectClickHouse(ctx context.Context, config *ClickHouseConfig) error {
	ctx, span := cm.tracer.Start(ctx, "ConnectClickHouse")
	defer span.End()

	log.Printf("Connecting to ClickHouse at %s:%d", config.Host, config.Port)

	// Create ClickHouse connection
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.DBName,
			Username: config.User,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug: false,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	cm.clickHouse = conn
	log.Printf("Successfully connected to ClickHouse")
	return nil
}

// GetYugabytePool returns the Yugabyte connection pool
func (cm *ConnectionManager) GetYugabytePool() *pgxpool.Pool {
	return cm.yugabytePool
}

// GetClickHouse returns the ClickHouse connection
func (cm *ConnectionManager) GetClickHouse() clickhouse.Conn {
	return cm.clickHouse
}

// Close closes all database connections
func (cm *ConnectionManager) Close() {
	if cm.yugabytePool != nil {
		cm.yugabytePool.Close()
		log.Println("Closed Yugabyte DB connection pool")
	}

	if cm.clickHouse != nil {
		if err := cm.clickHouse.Close(); err != nil {
			log.Printf("Error closing ClickHouse connection: %v", err)
		} else {
			log.Println("Closed ClickHouse connection")
		}
	}
}

// HealthCheck checks the health of all database connections
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	ctx, span := cm.tracer.Start(ctx, "DatabaseHealthCheck")
	defer span.End()

	// Check Yugabyte
	if cm.yugabytePool != nil {
		if err := cm.yugabytePool.Ping(ctx); err != nil {
			return fmt.Errorf("Yugabyte DB health check failed: %w", err)
		}
	}

	// Check ClickHouse
	if cm.clickHouse != nil {
		if err := cm.clickHouse.Ping(ctx); err != nil {
			return fmt.Errorf("ClickHouse health check failed: %w", err)
		}
	}

	return nil
}
