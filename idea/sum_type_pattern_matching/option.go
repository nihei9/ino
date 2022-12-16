package main

// in Ino:
//     data Option a = Some a
//                   | None

type tagID_Option string

const (
	tagID_Option_Some tagID_Option = "Some"
	tagID_Option_None tagID_Option = "None"
)

type Option[T Eqer] interface {
	Eqer
	Maybe() *matcher_Option[T]

	tag() tagID_Option
}

type tag_Option_Some[T Eqer] struct {
	p1 T
}

func Some[T Eqer](p1 T) Option[T] {
	return &tag_Option_Some[T]{
		p1: p1,
	}
}

func (x *tag_Option_Some[T]) Maybe() *matcher_Option[T] {
	return &matcher_Option[T]{
		x: x,
	}
}

func (x *tag_Option_Some[T]) Eq(y Eqer) bool {
	if z, ok := y.(*tag_Option_Some[T]); ok {
		return x.p1.Eq(z.p1)
	}
	return false
}

//nolint
func (x *tag_Option_Some[T]) tag() tagID_Option {
	return tagID_Option_Some
}

type tag_Option_None[T Eqer] struct {
}

func None[T Eqer]() Option[T] {
	return &tag_Option_None[T]{}
}

func (x *tag_Option_None[T]) Maybe() *matcher_Option[T] {
	return &matcher_Option[T]{
		x: x,
	}
}

func (x *tag_Option_None[T]) Eq(y Eqer) bool {
	_, ok := y.(*tag_Option_None[T])
	return ok
}

//nolint
func (x *tag_Option_None[T]) tag() tagID_Option {
	return tagID_Option_None
}

type matcher_Option[T Eqer] struct {
	x Option[T]
}

func (m *matcher_Option[T]) Some() *matcher_Option_Some[T] {
	if x, ok := m.x.(*tag_Option_Some[T]); ok {
		return &matcher_Option_Some[T]{
			x: x,
		}
	}
	return &matcher_Option_Some[T]{}
}

func (m *matcher_Option[T]) None() *matcher_Option_None[T] {
	if x, ok := m.x.(*tag_Option_None[T]); ok {
		return &matcher_Option_None[T]{
			x: x,
		}
	}
	return &matcher_Option_None[T]{}
}

type matcher_Option_Some[T Eqer] struct {
	x *tag_Option_Some[T]
}

func (x *matcher_Option_Some[T]) OK() bool {
	return x.x != nil
}

func (x *matcher_Option_Some[T]) Properties() (p1 T, ok bool) {
	if x.OK() {
		return x.x.p1, true
	}
	return
}

type matcher_Option_None[T Eqer] struct {
	x *tag_Option_None[T]
}

func (x *matcher_Option_None[T]) OK() bool {
	return x.x != nil
}
