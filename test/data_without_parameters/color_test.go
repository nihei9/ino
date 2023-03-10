//go:generate ino --package test --debug
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

	{
		cases, err := NewColorCaseSet(
			CaseRed(Red(), func() string {
				return "#ff0000"
			}),
			CaseGreen(Green(), func() string {
				return "#00ff00"
			}),
			CaseBlue(Blue(), func() string {
				return "#0000ff"
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		result, err := cases.Match(r)
		if err != nil || result != "#ff0000" {
			t.Errorf("unexpected matching result: %v, %v", result, err)
		}
	}

	{
		cases, err := NewColorCaseSet(
			CaseRed(Red(), func() string {
				return "apple"
			}),
			CaseGreen(Green(), func() string {
				return "green apple"
			}),
			CaseColorDefault(func(c Color) string {
				return "?"
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		apple, err := cases.Match(Red())
		if err != nil || apple != "apple" {
			t.Errorf("unexpected matching result: %v, %v", apple, err)
		}
		apple, err = cases.Match(Green())
		if err != nil || apple != "green apple" {
			t.Errorf("unexpected matching result: %v, %v", apple, err)
		}
		apple, err = cases.Match(Blue())
		if err != nil || apple != "?" {
			t.Errorf("unexpected matching result: %v, %v", apple, err)
		}
	}

	{
		cases, err := NewColorCaseSet(
			CaseColorDefault(func(c Color) string {
				return "?"
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		apple, err := cases.Match(Red())
		if err != nil || apple != "?" {
			t.Errorf("unexpected matching result: %v, %v", apple, err)
		}
	}

	{
		cases, err := NewColorCaseSet(
			CaseColorDefault(func(c Color) string {
				return "?"
			}),
			CaseRed(Red(), func() string {
				return "apple"
			}),
		)
		if err == nil || cases != nil {
			t.Error("error must occure")
		}
	}

	{
		cases, err := NewColorCaseSet(
			CaseColorDefault(func(c Color) string {
				return "?"
			}),
			CaseColorDefault(func(c Color) string {
				return "?"
			}),
		)
		if err == nil || cases != nil {
			t.Error("error must occure")
		}
	}
}
