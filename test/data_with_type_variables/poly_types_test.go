//go:generate ino --package test --debug
package test

import (
	"strconv"
	"strings"
	"testing"
)

func TestOption(t *testing.T) {
	s100 := Some(Int(100))
	ss100 := Some(s100)
	n := None[Int]()
	if !s100.Maybe().Some().OK() || s100.Maybe().None().OK() {
		t.Error("`s100` must be a Some")
	}
	if v, ok := s100.Maybe().Some().Fields(); !ok || v != 100 {
		t.Error("unexpected values")
	}
	if !ss100.Maybe().Some().OK() || ss100.Maybe().None().OK() {
		t.Error("`ss100` must be a Some")
	}
	if v, ok := ss100.Maybe().Some().Fields(); !ok || !v.Maybe().Some().OK() {
		t.Error("unexpected values")
	}
	if n.Maybe().Some().OK() || !n.Maybe().None().OK() {
		t.Error("`n` must be a Some")
	}

	v, ok := ApplyToSome(ss100, func(f1 Option[Int]) string {
		v, ok := ApplyToSome(s100, func(f1 Int) int {
			return int(f1) * 2
		})
		if !ok {
			return "!ok"
		}
		return strconv.Itoa(v)
	})
	if !ok || v != "200" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
	_, ok = ApplyToSome(n, func(f1 Int) string {
		return "ok"
	})
	if ok {
		t.Error("`n` must be None, not Some")
	}

	v, ok = ApplyToNone(n, func() string {
		return "ok"
	})
	if !ok || v != "ok" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
	_, ok = ApplyToNone(s100, func() string {
		return "ok"
	})
	if ok {
		t.Errorf("unexpected result: _, %v", ok)
	}

	cases, err := NewOptionCaseSet(
		CaseSome(Some(Int(100)), func(v Int) string {
			return strconv.Itoa(int(v))
		}),
		CaseNone(None[Int](), func() string {
			return "None"
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := cases.Match(s100)
	if err != nil || result != "100" {
		t.Errorf("unexpected matching result: %v, %v", result, err)
	}
}

func TestList(t *testing.T) {
	n := Nil[String]()
	l := Cons("!", n)
	l = Cons("world", l)
	l = Cons("Hello", l)

	if v := head(l); v != "Hello" {
		t.Errorf(`want: "Hello", got: %v`, strconv.Quote(string(v)))
	}
	tl := tail(l)
	if v := head(tl); v != "world" {
		t.Errorf(`want: "world", got: %v`, strconv.Quote(string(v)))
	}
	tl = tail(tl)
	if v := head(tl); v != "!" {
		t.Errorf(`want: "!", got: %v`, strconv.Quote(string(v)))
	}
	tl = tail(tl)
	if !tl.Maybe().Nil().OK() {
		t.Error("list is not Nil")
	}

	v, ok := ApplyToCons(l, func(h String, tl List[String]) string {
		v, _ := ApplyToCons(tl, func(h String, tl List[String]) string {
			v, _ := ApplyToCons(tl, func(h String, tl List[String]) string {
				v, _ := ApplyToNil(tl, func() string {
					return "🐱"
				})
				return strings.Repeat(string(h), 4) + " " + v
			})
			return strings.ToUpper(string(h)) + " " + v
		})
		return strings.ToUpper(string(h)) + " " + v
	})
	if !ok || v != "HELLO WORLD !!!! 🐱" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
}

func head[T Eqer](l List[T]) T {
	if e, _, ok := l.Maybe().Cons().Fields(); ok {
		return e
	}
	panic("list is Nil")
}

func tail[T Eqer](l List[T]) List[T] {
	if _, t, ok := l.Maybe().Cons().Fields(); ok {
		return t
	}
	panic("list is Nil")
}
