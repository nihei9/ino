package main

import "fmt"

func ApplyToSome[T Eqer, U any](x Option[T], f func(T) U) (U, bool) {
	p1, ok := x.Maybe().Some().Properties()
	if !ok {
		var zero U
		return zero, false
	}
	return f(p1), true
}

func ApplyToNone[T Eqer, U any](x Option[T], f func() U) (U, bool) {
	if !x.Maybe().None().OK() {
		var zero U
		return zero, false
	}
	return f(), true
}

var nonExhaustivePatternsErr = fmt.Errorf("non-exhaustive patterns")

func Match[T Eqer, U any](x Option[T], cases ...*optionCase[T, U]) (U, error) {
	var zero U
	for _, c := range cases {
		if c.err != nil {
			return zero, c.err
		}
	}
	for _, c := range cases {
		if result, ok := c.match(x); ok {
			return result, nil
		}
	}
	return zero, nonExhaustivePatternsErr
}

type optionCase[T Eqer, U any] struct {
	match func(Option[T]) (U, bool)
	err   error
}

func CaseSome[T Eqer, U any](y Option[T], f func(T) U) *optionCase[T, U] {
	var err error
	if ok := y.Maybe().Some().OK(); !ok {
		err = fmt.Errorf("condition must be a Some but given %v", y.tag())
	}
	return &optionCase[T, U]{
		match: func(x Option[T]) (U, bool) {
			if x.Eq(y) {
				result, _ := ApplyToSome(x, f)
				return result, true
			}
			var zero U
			return zero, false
		},
		err: err,
	}
}

func CaseNone[T Eqer, U any](y Option[T], f func() U) *optionCase[T, U] {
	var err error
	if ok := y.Maybe().None().OK(); !ok {
		err = fmt.Errorf("condition must be a None but given %v", y.tag())
	}
	return &optionCase[T, U]{
		match: func(x Option[T]) (U, bool) {
			if x.Eq(y) {
				result, _ := ApplyToNone(x, f)
				return result, true
			}
			var zero U
			return zero, false
		},
		err: err,
	}
}
