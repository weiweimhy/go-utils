package envutil

import (
	"reflect"
	"testing"
	"time"
)

func TestTypedEnv(t *testing.T) {
	t.Setenv("APP_NAME", "api")
	t.Setenv("APP_DEBUG", "true")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_TIMEOUT", "2s")
	t.Setenv("APP_LIST", "a, b,,c")

	if got := String("APP_NAME", "fallback"); got != "api" {
		t.Fatalf("String() = %q", got)
	}
	if got := Bool("APP_DEBUG", false); !got {
		t.Fatal("Bool() = false")
	}
	if got := Int("APP_PORT", 80); got != 8080 {
		t.Fatalf("Int() = %d", got)
	}
	if got := Duration("APP_TIMEOUT", time.Second); got != 2*time.Second {
		t.Fatalf("Duration() = %v", got)
	}
	if got := List("APP_LIST", ",", nil); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("List() = %#v", got)
	}
}

func TestFallbacks(t *testing.T) {
	t.Setenv("BAD_BOOL", "maybe")
	if got := Bool("BAD_BOOL", true); !got {
		t.Fatal("Bool() should return fallback")
	}
	if got := Int("MISSING", 42); got != 42 {
		t.Fatalf("Int() = %d", got)
	}
}
