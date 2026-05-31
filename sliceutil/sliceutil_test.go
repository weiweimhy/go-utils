package sliceutil

import (
	"reflect"
	"testing"
)

func TestChunk(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunk() = %#v", got)
	}
}

func TestUniqueMapFilter(t *testing.T) {
	unique := Unique([]int{1, 2, 1, 3})
	if !reflect.DeepEqual(unique, []int{1, 2, 3}) {
		t.Fatalf("Unique() = %#v", unique)
	}
	mapped := Map(unique, func(v int) int { return v * 2 })
	filtered := Filter(mapped, func(v int) bool { return v > 2 })
	if !reflect.DeepEqual(filtered, []int{4, 6}) {
		t.Fatalf("filtered = %#v", filtered)
	}
}
