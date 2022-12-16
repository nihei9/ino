//go:generate ino --package test
package test

import "testing"

func TestColor(t *testing.T) {
	b := BallpointPen(1, Black())
	size, color, ok := b.Maybe().BallpointPen().Fields()
	if !ok {
		t.Error("b must be a BallpointPen")
	}
	if size != 1 || !color.Maybe().Black().OK() {
		t.Error("unexpected values")
	}
}
