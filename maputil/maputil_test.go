package maputil

import (
	"reflect"
	"testing"
)

func TestKeysSortedCloneMerge(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	if got := KeysSorted(m); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("KeysSorted() = %#v", got)
	}
	clone := Clone(m)
	clone["a"] = 10
	if m["a"] != 1 {
		t.Fatal("Clone() should be shallow but independent")
	}
	merged := Merge(m, map[string]int{"a": 3, "c": 4})
	if !reflect.DeepEqual(merged, map[string]int{"a": 3, "b": 2, "c": 4}) {
		t.Fatalf("Merge() = %#v", merged)
	}
}
