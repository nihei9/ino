//go:generate ino --package test
package test

import "testing"

func TestOption(t *testing.T) {
	s100 := Some(100)
	ss100 := Some(s100)
	n := None[int]()
	if !s100.Maybe().Some().OK() || s100.Maybe().None().OK() {
		t.Error("s100 must be a Some")
	}
	if v, ok := s100.Maybe().Some().Fields(); !ok || v != 100 {
		t.Error("unexpected values")
	}
	if !ss100.Maybe().Some().OK() || ss100.Maybe().None().OK() {
		t.Error("ss100 must be a Some")
	}
	if v, ok := ss100.Maybe().Some().Fields(); !ok || !v.Maybe().Some().OK() {
		t.Error("unexpected values")
	}
	if n.Maybe().Some().OK() || !n.Maybe().None().OK() {
		t.Error("n must be a Some")
	}
}
