package mongo

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	cfg := Config{
		URI:              "mongodb://localhost:27017",
		DatabaseName:     "testdb",
		ConnectTimeout:   5 * time.Second,
		OperationTimeout: 5 * time.Second,
	}

	if cfg.URI != "mongodb://localhost:27017" {
		t.Errorf("expected URI mongodb://localhost:27017, got %s", cfg.URI)
	}
}
