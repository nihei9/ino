//go:generate ino --package test
package test

import "testing"

func TestColor(t *testing.T) {
	r := Red()
	g := Green()
	b := Blue()

	if !r.Maybe().Red().OK() {
		t.Error("r must be Red")
	}
	if r.Maybe().Green().OK() {
		t.Error("r isn't Green")
	}
	if r.Maybe().Blue().OK() {
		t.Error("r isn't Blue")
	}

	if g.Maybe().Red().OK() {
		t.Error("g isn't Red")
	}
	if !g.Maybe().Green().OK() {
		t.Error("g must be Green")
	}
	if g.Maybe().Blue().OK() {
		t.Error("g isn't Blue")
	}

	if b.Maybe().Red().OK() {
		t.Error("b isn't Red")
	}
	if b.Maybe().Green().OK() {
		t.Error("b isn't Green")
	}
	if !b.Maybe().Blue().OK() {
		t.Error("b must be Blue")
	}
}
