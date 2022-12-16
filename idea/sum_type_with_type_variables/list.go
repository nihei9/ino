package main

// in Ino:
//     data List a = Cons a (List a)
//                 | Nil

type List[T any] interface {
	Match() *matcher_List[T]
}

type tag_List_Cons[T any] struct {
	p1 T
	p2 List[T]
}

func Cons[T any](p1 T, p2 List[T]) List[T] {
	return &tag_List_Cons[T]{
		p1: p1,
		p2: p2,
	}
}

func (x *tag_List_Cons[T]) Match() *matcher_List[T] {
	return &matcher_List[T]{
		x: x,
	}
}

type tag_List_Nil[T any] struct {
}

func Nil[T any]() List[T] {
	return &tag_List_Nil[T]{}
}

func (x *tag_List_Nil[T]) Match() *matcher_List[T] {
	return &matcher_List[T]{
		x: x,
	}
}

type matcher_List[T any] struct {
	x List[T]
}

func (m *matcher_List[T]) AsCons() *maybe_tag_List_Cons[T] {
	if x, ok := m.x.(*tag_List_Cons[T]); ok {
		return &maybe_tag_List_Cons[T]{
			x: x,
		}
	}
	return &maybe_tag_List_Cons[T]{}
}

func (m *matcher_List[T]) AsNil() *maybe_tag_List_Nil[T] {
	if x, ok := m.x.(*tag_List_Nil[T]); ok {
		return &maybe_tag_List_Nil[T]{
			x: x,
		}
	}
	return &maybe_tag_List_Nil[T]{}
}

type maybe_tag_List_Cons[T any] struct {
	x *tag_List_Cons[T]
}

func (x *maybe_tag_List_Cons[T]) OK() bool {
	return x.x != nil
}

func (x *maybe_tag_List_Cons[T]) Parameters() (p1 T, p2 List[T], ok bool) {
	if x.OK() {
		return x.x.p1, x.x.p2, true
	}
	return
}

type maybe_tag_List_Nil[T any] struct {
	x *tag_List_Nil[T]
}

func (x *maybe_tag_List_Nil[T]) OK() bool {
	return x.x != nil
}
