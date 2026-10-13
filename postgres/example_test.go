package postgres_test

import (
	"context"
	"os"

	"github.com/weiweimhy/go-utils/postgres/v6"
)

func ExampleNewPool() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return
	}

	pool, err := postgres.NewPool(ctx, postgres.Config{ConnString: databaseURL})
	if err != nil {
		return
	}
	defer pool.Close()

	// Reuse pool for the application lifecycle and pass it to request handlers.
}
