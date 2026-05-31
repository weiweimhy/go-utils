package jsonfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func TestLoadMissingOK(t *testing.T) {
	fallback := testConfig{Name: "fallback", Port: 8080}
	got, err := Load(filepath.Join(t.TempDir(), "missing.json"), fallback, Options{MissingOK: true})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != fallback {
		t.Fatalf("Load() = %+v, want %+v", got, fallback)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "app.json")
	want := testConfig{Name: "api", Port: 8080}

	if err := Save(path, want, Options{AtomicSave: true, Sync: true}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(raw), "\n  \"name\"") {
		t.Fatalf("expected indented JSON, got %q", string(raw))
	}

	got, err := Load(path, testConfig{}, Options{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}
