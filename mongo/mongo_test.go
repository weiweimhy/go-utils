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

func TestNewClientAcceptsNilContext(t *testing.T) {
	config := DefaultConfig()
	config.URI = "mongodb://127.0.0.1:1"
	config.ConnectTimeout = time.Millisecond

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("NewClient(nil, ...) panicked: %v", recovered)
		}
	}()
	_, _ = NewClient(nil, config)
}

func TestOperationContextAcceptsNilContext(t *testing.T) {
	client := &Client{cfg: DefaultConfig()}
	ctx, cancel := client.operationContext(nil)
	defer cancel()
	if ctx == nil {
		t.Fatal("operation context should not be nil")
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("operation context error = %v", err)
	}
}
