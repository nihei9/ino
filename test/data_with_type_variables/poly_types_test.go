//go:generate ino --package test --debug
package test

import (
	"strconv"
	"strings"
	"testing"
)

func TestOption(t *testing.T) {
	s100 := Some(100)
	ss100 := Some(s100)
	n := None[int]()
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

	v, ok := ApplyToSome(ss100, func(f1 Option[int]) string {
		v, ok := ApplyToSome(s100, func(f1 int) int {
			return f1 * 2
		})
		if !ok {
			return "!ok"
		}
		return strconv.Itoa(v)
	})
	if !ok || v != "200" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
	_, ok = ApplyToSome(n, func(f1 int) string {
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
}

func TestList(t *testing.T) {
	n := Nil[string]()
	l := Cons("!", n)
	l = Cons("world", l)
	l = Cons("Hello", l)

	if v := head(l); v != "Hello" {
		t.Errorf(`want: "Hello", got: %v`, strconv.Quote(v))
	}
	tl := tail(l)
	if v := head(tl); v != "world" {
		t.Errorf(`want: "world", got: %v`, strconv.Quote(v))
	}
	tl = tail(tl)
	if v := head(tl); v != "!" {
		t.Errorf(`want: "!", got: %v`, strconv.Quote(v))
	}
	tl = tail(tl)
	if !tl.Maybe().Nil().OK() {
		t.Error("list is not Nil")
	}

	v, ok := ApplyToCons(l, func(h string, tl List[string]) string {
		v, _ := ApplyToCons(tl, func(h string, tl List[string]) string {
			v, _ := ApplyToCons(tl, func(h string, tl List[string]) string {
				v, _ := ApplyToNil(tl, func() string {
					return "🐱"
				})
				return strings.Repeat(h, 4) + " " + v
			})
			return strings.ToUpper(h) + " " + v
		})
		return strings.ToUpper(h) + " " + v
	})
	if !ok || v != "HELLO WORLD !!!! 🐱" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
}

func head[T any](l List[T]) T {
	if e, _, ok := l.Maybe().Cons().Fields(); ok {
		return e
	}
	panic("list is Nil")
}

func tail[T any](l List[T]) List[T] {
	if _, t, ok := l.Maybe().Cons().Fields(); ok {
		return t
	}
	panic("list is Nil")
}
