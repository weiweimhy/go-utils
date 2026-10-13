package postgres

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPoolWithConnectionString(t *testing.T) {
	isolatePGEnvironment(t)
	server := startMockPostgres(t)
	dsn := "postgres://app-user:private-password@" + server.address + "/app-db?sslmode=disable&pool_max_conns=1"

	pool, err := NewPool(context.Background(), Config{ConnString: dsn})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	if got := pool.Config().MaxConns; got != 1 {
		t.Errorf("pool max connections = %d, want 1", got)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Errorf("Ping() after NewPool() returned: %v", err)
	}
	var answer int
	if err := pool.QueryRow(context.Background(), "select 1").Scan(&answer); err != nil {
		t.Errorf("query from returned pool: %v", err)
	} else if answer != 1 {
		t.Errorf("query result = %d, want 1", answer)
	}
	pool.Close()
	pool.Close()
	server.waitForClose(t)
}

func TestPoolCloseWaitsForAcquiredConnection(t *testing.T) {
	isolatePGEnvironment(t)
	server := startMockPostgres(t)
	pool, err := NewPool(context.Background(), Config{ConnString: "postgres://user:password@" + server.address + "/db?sslmode=disable"})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		pool.Close()
		t.Fatalf("Acquire() error = %v", err)
	}
	released := false
	defer func() {
		if !released {
			conn.Release()
		}
	}()

	closeStarted := make(chan struct{})
	closeDone := make(chan struct{})
	go func() {
		close(closeStarted)
		pool.Close()
		close(closeDone)
	}()
	<-closeStarted
	select {
	case <-closeDone:
		t.Fatal("Pool.Close() returned while a connection was acquired")
	case <-time.After(100 * time.Millisecond):
	}

	conn.Release()
	released = true
	select {
	case <-closeDone:
	case <-time.After(3 * time.Second):
		t.Fatal("Pool.Close() did not return after the connection was released")
	}
	server.waitForClose(t)
}

func TestNewPoolWithEmptyConnectionStringUsesEnvironment(t *testing.T) {
	isolatePGEnvironment(t)
	server := startMockPostgres(t)
	host, port, err := net.SplitHostPort(server.address)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PGHOST", host)
	t.Setenv("PGPORT", port)
	t.Setenv("PGDATABASE", "environment-db")
	t.Setenv("PGUSER", "environment-user")
	t.Setenv("PGPASSWORD", "environment-password")
	t.Setenv("PGSSLMODE", "disable")

	pool, err := NewPool(context.Background(), Config{})
	if err != nil {
		t.Fatalf("NewPool() with environment configuration error = %v", err)
	}
	pool.Close()
	server.waitForClose(t)
}

func TestNewPoolRejectsNilContext(t *testing.T) {
	isolatePGEnvironment(t)
	pool, err := NewPool(nil, Config{})
	if pool != nil {
		t.Fatal("NewPool() returned a pool for a nil context")
	}
	if !errors.Is(err, ErrNilContext) {
		t.Errorf("NewPool() error = %v, want errors.Is(err, ErrNilContext)", err)
	}
}

func TestNewPoolRejectsNegativeStartupTimeout(t *testing.T) {
	isolatePGEnvironment(t)
	pool, err := NewPool(context.Background(), Config{StartupTimeout: -time.Nanosecond})
	if pool != nil {
		t.Fatal("NewPool() returned a pool for a negative startup timeout")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("NewPool() error = %v, want errors.Is(err, ErrInvalidConfig)", err)
	}
}

func TestNewPoolRejectsInvalidConnectionString(t *testing.T) {
	isolatePGEnvironment(t)
	pool, err := NewPool(context.Background(), Config{ConnString: "postgres://user:secret@%zz/db"})
	if pool != nil {
		t.Fatal("NewPool() returned a pool for an invalid connection string")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("NewPool() error = %v, want errors.Is(err, ErrInvalidConfig)", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Errorf("NewPool() error exposed connection string data: %v", err)
	}
}

func TestNewPoolRejectsInvalidPoolOptions(t *testing.T) {
	isolatePGEnvironment(t)
	tests := []struct {
		name       string
		parameters string
	}{
		{name: "zero health check period", parameters: "pool_health_check_period=0s"},
		{name: "negative health check period", parameters: "pool_health_check_period=-1s"},
		{name: "negative minimum connections", parameters: "pool_min_conns=-1"},
		{name: "negative minimum idle connections", parameters: "pool_min_idle_conns=-1"},
		{name: "minimum connections exceed maximum", parameters: "pool_max_conns=1&pool_min_conns=2"},
		{name: "minimum idle connections exceed maximum", parameters: "pool_max_conns=1&pool_min_idle_conns=2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := "postgres://user:password@localhost/db?sslmode=disable&" + test.parameters
			pool, err := NewPool(context.Background(), Config{ConnString: dsn})
			if pool != nil {
				pool.Close()
				t.Fatal("NewPool() returned a pool for invalid pool options")
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("NewPool() error = %v, want errors.Is(err, ErrInvalidConfig)", err)
			}
		})
	}
}

func TestNewPoolConnectionFailureIsSanitizedAndInspectable(t *testing.T) {
	isolatePGEnvironment(t)
	address := unusedTCPAddress(t)
	dsn := "postgres://private-user:private-password@" + address + "/private-database?sslmode=disable"
	pool, err := NewPool(context.Background(), Config{
		ConnString:     dsn,
		StartupTimeout: 3 * time.Second,
	})
	if pool != nil {
		t.Fatal("NewPool() returned a pool after connection failure")
	}
	if err == nil {
		t.Fatal("NewPool() error = nil, want connection failure")
	}
	for _, secret := range []string{"private-user", "private-password", "private-database"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("NewPool() error exposed %q: %v", secret, err)
		}
	}
	var networkError net.Error
	if !errors.As(err, &networkError) {
		t.Errorf("NewPool() error does not preserve the underlying network error type: %T", err)
	}
}

func TestNewPoolStartupTimeoutClosesConnection(t *testing.T) {
	isolatePGEnvironment(t)
	server := startStalledPostgres(t)
	pool, err := NewPool(context.Background(), Config{
		ConnString:     server.connectionString,
		StartupTimeout: 300 * time.Millisecond,
	})
	if pool != nil {
		t.Fatal("NewPool() returned a pool after startup timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("NewPool() error = %v, want context.DeadlineExceeded", err)
	}
	server.waitForStartup(t)
	server.waitForClose(t)
}

func TestNewPoolPingFailurePreservesPgErrorAndClosesConnection(t *testing.T) {
	isolatePGEnvironment(t)
	server := startFailingMockPostgres(t)
	dsn := "postgres://private-user:private-password@" + server.address + "/private-database?sslmode=disable"
	pool, err := NewPool(context.Background(), Config{ConnString: dsn})
	if pool != nil {
		t.Fatal("NewPool() returned a pool after Ping failure")
	}
	if err == nil {
		t.Fatal("NewPool() error = nil, want Ping failure")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("NewPool() error does not preserve pgconn.PgError: %T", err)
	}
	if pgErr.Message != "mock Ping failure" {
		t.Errorf("PgError message = %q, want %q", pgErr.Message, "mock Ping failure")
	}
	for _, secret := range []string{"private-user", "private-password", "private-database"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("NewPool() error exposed %q: %v", secret, err)
		}
	}
	server.waitForClose(t)
}

func TestNewPoolCancellationClosesConnection(t *testing.T) {
	isolatePGEnvironment(t)
	server := startStalledPostgres(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type result struct {
		pool *pgxpool.Pool
		err  error
	}
	completed := make(chan result, 1)
	go func() {
		pool, err := NewPool(ctx, Config{
			ConnString:     server.connectionString,
			StartupTimeout: 5 * time.Second,
		})
		completed <- result{pool: pool, err: err}
	}()

	server.waitForStartup(t)
	cancel()
	select {
	case got := <-completed:
		if got.pool != nil {
			got.pool.Close()
			t.Fatal("NewPool() returned a pool after cancellation")
		}
		if !errors.Is(got.err, context.Canceled) {
			t.Errorf("NewPool() error = %v, want context.Canceled", got.err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("NewPool() did not return after context cancellation")
	}
	server.waitForClose(t)
}

func isolatePGEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"PGHOST",
		"PGHOSTADDR",
		"PGPORT",
		"PGDATABASE",
		"PGUSER",
		"PGPASSWORD",
		"PGPASSFILE",
		"PGSERVICE",
		"PGSERVICEFILE",
		"PGSSLMODE",
		"PGSSLCERT",
		"PGSSLKEY",
		"PGSSLROOTCERT",
		"PGSSLPASSWORD",
		"PGOPTIONS",
		"PGAPPNAME",
		"PGCONNECT_TIMEOUT",
		"PGTARGETSESSIONATTRS",
		"PGTZ",
		"PGMINPROTOCOLVERSION",
		"PGMAXPROTOCOLVERSION",
		"PGSSLSNI",
		"PGSSLNEGOTIATION",
		"PGCHANNELBINDING",
		"PGREQUIREAUTH",
	} {
		t.Setenv(name, "")
	}
}
