//go:generate ino --package test --debug
package test

import (
	"strconv"
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
}

func TestList(t *testing.T) {
	n := Nil[string]()
	l := Cons("!", n)
	l = Cons("world", l)
	l = Cons("Hello", l)

	if v := head(l); v != "Hello" {
		t.Errorf(`want: "Hello", got: %v`, strconv.Quote(v))
	}
	l = tail(l)
	if v := head(l); v != "world" {
		t.Errorf(`want: "world", got: %v`, strconv.Quote(v))
	}
	l = tail(l)
	if v := head(l); v != "!" {
		t.Errorf(`want: "!", got: %v`, strconv.Quote(v))
	}
	l = tail(l)
	if !l.Maybe().Nil().OK() {
		t.Error("list is not Nil")
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
