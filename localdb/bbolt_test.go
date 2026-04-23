package localdb

import "testing"

func TestOpenSetAndGet(t *testing.T) {
	db, err := Open(t.TempDir(), "test.db")
	if err != nil {
		t.Fatalf("expected open to succeed, got %v", err)
	}
	defer db.Close()

	if err := db.Set("bucket", "key", []byte("value")); err != nil {
		t.Fatalf("expected set to succeed, got %v", err)
	}

	got, err := db.Get("bucket", "key")
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}
	if string(got) != "value" {
		t.Fatalf("expected value, got %q", string(got))
	}
}

func TestSetAndGetJSON(t *testing.T) {
	db, err := Open(t.TempDir(), "test.db")
	if err != nil {
		t.Fatalf("expected open to succeed, got %v", err)
	}
	defer db.Close()

	type user struct {
		Name string `json:"name"`
	}

	want := user{Name: "alice"}
	if err := db.SetJSON("users", "1", want); err != nil {
		t.Fatalf("expected SetJSON to succeed, got %v", err)
	}

	var got user
	if err := db.GetJSON("users", "1", &got); err != nil {
		t.Fatalf("expected GetJSON to succeed, got %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}
