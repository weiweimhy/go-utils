package jsonutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/weiweimhy/go-utils/v6/streamutil"
)

type strictConfig struct {
	Name   string `json:"name"`
	Nested struct {
		Enabled bool `json:"enabled"`
	} `json:"nested"`
}

func TestDecodeStrict(t *testing.T) {
	got, err := DecodeStrict[strictConfig](strings.NewReader(`{"name":"alice","nested":{"enabled":true}}`), 1024)
	if err != nil {
		t.Fatalf("DecodeStrict() error = %v", err)
	}
	if got.Name != "alice" || !got.Nested.Enabled {
		t.Fatalf("DecodeStrict() = %+v", got)
	}
}

func TestDecodeStrictRejectsUnknownDuplicateAndTrailingValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unknown field", input: `{"name":"alice","extra":true}`},
		{name: "duplicate root field", input: `{"name":"alice","name":"bob"}`},
		{name: "duplicate nested field", input: `{"name":"alice","nested":{"enabled":true,"enabled":false}}`},
		{name: "trailing value", input: `{"name":"alice"} {}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeStrict[strictConfig](strings.NewReader(tt.input), 1024)
			if err == nil {
				t.Fatal("DecodeStrict() error = nil, want an error")
			}
		})
	}
}

func TestDecodeStrictLimit(t *testing.T) {
	_, err := DecodeStrict[strictConfig](strings.NewReader(`{"name":"alice"}`), 4)
	if !errors.Is(err, streamutil.ErrLimitExceeded) {
		t.Fatalf("DecodeStrict() error = %v, want ErrLimitExceeded", err)
	}
}
