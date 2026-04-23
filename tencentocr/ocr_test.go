package tencentocr

import "testing"

func TestNewClient(t *testing.T) {
	client, err := NewClient(Config{
		SecretId:  "id",
		SecretKey: "key",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}
