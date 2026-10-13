package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationPostgres(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := NewPool(ctx, Config{ConnString: dsn})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	var value int
	if err := pool.QueryRow(ctx, "select 1").Scan(&value); err != nil {
		t.Fatalf("query constant: %v", err)
	}
	if value != 1 {
		t.Fatalf("query result = %d, want 1", value)
	}
}
