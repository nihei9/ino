package main

// in Ino:
//     data Option a = Some a
//                   | None

type Option[T any] interface {
	Match() *matcher_Option[T]
}

type tag_Option_Some[T any] struct {
	p1 T
}

func Some[T any](p1 T) Option[T] {
	return &tag_Option_Some[T]{
		p1: p1,
	}
}

func (x *tag_Option_Some[T]) Match() *matcher_Option[T] {
	return &matcher_Option[T]{
		x: x,
	}
}

type tag_Option_None[T any] struct {
}

func None[T any]() Option[T] {
	return &tag_Option_None[T]{}
}

func (x *tag_Option_None[T]) Match() *matcher_Option[T] {
	return &matcher_Option[T]{
		x: x,
	}
}

type matcher_Option[T any] struct {
	x Option[T]
}

func (m *matcher_Option[T]) AsSome() *maybe_tag_Option_Some[T] {
	if x, ok := m.x.(*tag_Option_Some[T]); ok {
		return &maybe_tag_Option_Some[T]{
			x: x,
		}
	}
	return &maybe_tag_Option_Some[T]{}
}

func (m *matcher_Option[T]) AsNone() *maybe_tag_Option_None[T] {
	if x, ok := m.x.(*tag_Option_None[T]); ok {
		return &maybe_tag_Option_None[T]{
			x: x,
		}
	}
	return &maybe_tag_Option_None[T]{}
}

type maybe_tag_Option_Some[T any] struct {
	x *tag_Option_Some[T]
}

func (x *maybe_tag_Option_Some[T]) OK() bool {
	return x.x != nil
}

func (x *maybe_tag_Option_Some[T]) Parameters() (p1 T, ok bool) {
	if x.OK() {
		return x.x.p1, true
	}
	return
}

type maybe_tag_Option_None[T any] struct {
	x *tag_Option_None[T]
}

func (x *maybe_tag_Option_None[T]) OK() bool {
	return x.x != nil
}
