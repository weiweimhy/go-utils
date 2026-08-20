package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetBytesFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	data, err := GetBytesFromURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("expected ok, got %q", string(data))
	}
}

func TestGetBytesFromURLStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := GetBytesFromURL(context.Background(), server.URL)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestDefaultSuccessStatusAcceptsAny2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	defer server.Close()

	data, err := GetBytes(context.Background(), server.URL, Options{})
	if err != nil || string(data) != "created" {
		t.Fatalf("GetBytes() = %q, %v", data, err)
	}
}

func TestStatusErrorRedactsURLAndDoesNotCaptureBodyByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private error response", http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL+"?access_token=secret", Options{})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("GetBytes() error = %v, want StatusError", err)
	}
	if strings.Contains(statusErr.URL, "secret") || len(statusErr.Body) != 0 {
		t.Fatalf("unsafe status error: %+v", statusErr)
	}
}

func TestStatusErrorCapturesLimitedBodyWhenRequested(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private error response", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL, Options{CaptureErrorBody: true, MaxErrorBodyBytes: 4})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("GetBytes() error = %v, want StatusError", err)
	}
	if len(statusErr.Body) != 0 {
		t.Fatalf("captured body = %q, want omitted body after limit", statusErr.Body)
	}
}

func TestGetBytesMaxBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("too large"))
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL, Options{MaxBytes: 3})
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestPostJSON(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		OK   bool   `json:"ok"`
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(response{OK: true, Name: req.Name})
	}))
	defer server.Close()

	got, err := PostJSON[response](context.Background(), server.URL, request{Name: "alice"}, Options{})
	if err != nil {
		t.Fatalf("PostJSON() error = %v", err)
	}
	if !got.OK || got.Name != "alice" {
		t.Fatalf("PostJSON() = %+v", got)
	}
}

func TestDoJSONDoesNotMutateCallerHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	headers := make(http.Header)
	headers.Set("X-Request-ID", "request-1")
	_, err := PostJSON[map[string]bool](context.Background(), server.URL, map[string]bool{"request": true}, Options{
		Headers: headers,
	})
	if err != nil {
		t.Fatalf("PostJSON() error = %v", err)
	}
	if got := headers.Get("Content-Type"); got != "" {
		t.Fatalf("caller Content-Type = %q, want empty", got)
	}
	if got := headers.Get("Accept"); got != "" {
		t.Fatalf("caller Accept = %q, want empty", got)
	}
	if got := headers.Get("X-Request-ID"); got != "request-1" {
		t.Fatalf("caller header changed: %q", got)
	}
}
