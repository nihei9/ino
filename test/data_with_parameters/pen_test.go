//go:generate ino --package test --debug
package test

import (
	"fmt"
	"testing"
)

func TestColor(t *testing.T) {
	b := BallpointPen(1, Black())
	size, color, ok := b.Maybe().BallpointPen().Fields()
	if !ok {
		t.Error("b must be a BallpointPen")
	}
	if size != 1 || !color.Maybe().Black().OK() {
		t.Error("unexpected values")
	}

	cases, err := NewPenCaseSet(
		CaseBallpointPen(BallpointPen(1, Black()), func(size Int, color Color) string {
			cases, err := NewColorCaseSet(
				CaseRed(Red(), func() string {
					return "Red"
				}),
				CaseGreen(Green(), func() string {
					return "Green"
				}),
				CaseBlue(Blue(), func() string {
					return "Blue"
				}),
				CaseBlack(Black(), func() string {
					return "Black"
				}),
			)
			if err != nil {
				t.Fatal(err)
			}
			c, err := cases.Match(color)
			if err != nil {
				t.Fatal(err)
			}
			return fmt.Sprintf("ballpoint-%v-%v", size, c)
		}),
		CaseFountainPen(FountainPen(), func() string {
			return "fountain"
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := cases.Match(b)
	if err != nil || result != "ballpoint-1-Black" {
		t.Errorf("unexpected matching result: %v, %v", result, err)
	}
}
