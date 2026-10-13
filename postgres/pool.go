package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultStartupTimeout = 10 * time.Second

// ErrInvalidConfig indicates that Config or a pool_* connection parameter is
// invalid, or that pgx could not parse the connection string.
var ErrInvalidConfig = errors.New("postgres: invalid configuration")

// ErrNilContext indicates that NewPool received a nil context.
var ErrNilContext = errors.New("postgres: nil context")

// Config controls pool creation and the initial connectivity check.
//
// ConnString uses pgx's connection string syntax. Pool options such as
// pool_max_conns and pool_min_conns are parsed by pgxpool. When ConnString is
// empty, pgx reads its usual PG* environment variables.
//
// StartupTimeout limits pool creation and the initial Ping only. A zero value
// uses ten seconds; a negative value is invalid. It does not limit later pool
// operations.
type Config struct {
	ConnString     string
	StartupTimeout time.Duration
}

// NewPool creates a pgxpool.Pool and verifies that it can connect to the
// database before returning it. The caller owns the returned pool and must call
// Pool.Close when finished.
//
// A nil context or negative StartupTimeout is rejected. If StartupTimeout is
// zero, it defaults to ten seconds. The startup timeout covers pool creation
// and Ping, not subsequent queries. If startup fails after the pool has been
// created, NewPool closes it before returning.
//
// Returned errors have sanitized Error text. The wrapped error remains
// available through errors.Is and errors.As; inspecting that underlying error
// may reveal connection details, including credentials from ConnString.
func NewPool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	if ctx == nil {
		return nil, &operationError{operation: "validate context", err: ErrNilContext}
	}
	if config.StartupTimeout < 0 {
		return nil, &operationError{operation: "validate configuration", err: ErrInvalidConfig}
	}

	poolConfig, err := pgxpool.ParseConfig(config.ConnString)
	if err != nil {
		return nil, &operationError{
			operation: "parse connection configuration",
			err:       errors.Join(ErrInvalidConfig, err),
		}
	}
	if err := validatePoolConfig(poolConfig); err != nil {
		return nil, err
	}

	startupTimeout := config.StartupTimeout
	if startupTimeout == 0 {
		startupTimeout = defaultStartupTimeout
	}
	startupCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(startupCtx, poolConfig)
	if err != nil {
		if pool != nil {
			pool.Close()
		}
		return nil, &operationError{operation: "create connection pool", err: err}
	}

	if err := pool.Ping(startupCtx); err != nil {
		pool.Close()
		return nil, &operationError{operation: "verify database connection", err: err}
	}

	return pool, nil
}

func validatePoolConfig(config *pgxpool.Config) error {
	switch {
	case config.HealthCheckPeriod <= 0:
		return &operationError{operation: "validate pool health check period", err: ErrInvalidConfig}
	case config.MinConns < 0:
		return &operationError{operation: "validate pool minimum connections", err: ErrInvalidConfig}
	case config.MinIdleConns < 0:
		return &operationError{operation: "validate pool minimum idle connections", err: ErrInvalidConfig}
	case config.MinConns > config.MaxConns:
		return &operationError{operation: "validate pool minimum connections", err: ErrInvalidConfig}
	case config.MinIdleConns > config.MaxConns:
		return &operationError{operation: "validate pool minimum idle connections", err: ErrInvalidConfig}
	default:
		return nil
	}
}

type operationError struct {
	operation string
	err       error
}

func (err *operationError) Error() string {
	return "postgres: " + err.operation + " failed"
}

func (err *operationError) Unwrap() error {
	return err.err
}
