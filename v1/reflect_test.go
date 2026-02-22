package logger

import (
	"reflect"
	"testing"
)

type testNested struct {
	A int
	B struct {
		C string
		D bool
	}
	E float64
}

func TestCollectFields(t *testing.T) {
	t1 := reflect.TypeOf(testNested{})
	fields := collectFields(t1, "")
	expected := []string{"A", "B.C", "B.D", "E"}
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("collectFields = %v, want %v", fields, expected)
	}
}

func TestGetFieldInfos(t *testing.T) {
	infos := getFieldInfos(reflect.TypeOf(testNested{}))
	if len(infos) != 4 {
		t.Fatalf("got %d fieldInfos, want 4", len(infos))
	}
	names := make([]string, len(infos))
	for i, fi := range infos {
		names[i] = fi.fullName
	}
	expectedNames := []string{"A", "B.C", "B.D", "E"}
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("fullNames = %v, want %v", names, expectedNames)
	}
	expectedIndices := [][]int{{0}, {1, 0}, {1, 1}, {2}}
	for i, fi := range infos {
		if !reflect.DeepEqual(fi.index, expectedIndices[i]) {
			t.Errorf("index for %s = %v, want %v", fi.fullName, fi.index, expectedIndices[i])
		}
	}
}

func TestGetFieldInfosCaching(t *testing.T) {
	t1 := reflect.TypeOf(testNested{})
	infos1 := getFieldInfos(t1)
	infos2 := getFieldInfos(t1)
	if len(infos1) != len(infos2) {
		t.Error("cached result length mismatch")
	}
	// The slices should be identical (same backing array) due to caching.
	if &infos1[0] != &infos2[0] {
		t.Error("cached slice not reused")
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		val  any
		want bool
	}{
		{0, true},
		{1, false},
		{"", true},
		{"hello", false},
		{struct{}{}, true},
		{struct{ A int }{0}, true},
		{struct{ A int }{1}, false},
		{[]int{}, false}, // nil slice is zero; empty non-nil slice is not zero according to reflect.IsZero
		{[]int{1}, false},
		{nil, true},
	}
	for _, tt := range tests {
		v := reflect.ValueOf(tt.val)
		got := isZeroValue(v)
		if got != tt.want {
			t.Errorf("isZeroValue(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}
