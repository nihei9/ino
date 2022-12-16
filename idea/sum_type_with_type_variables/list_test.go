package main

import "fmt"

func ExampleList() {
	// in Ino:
	//     l = Cons "bar" (Cons "foo" Nil)
	l := Nil[string]()
	l = Cons("foo", l)
	l = Cons("bar", l)

	head, tail, ok := l.Match().AsCons().Parameters()
	if ok {
		fmt.Println(head)
	}
	head, tail, ok = tail.Match().AsCons().Parameters()
	if ok {
		fmt.Println(head)
	}
	if tail.Match().AsNil().OK() {
		fmt.Println("Nil")
	}

	// Output:
	// bar
	// foo
	// Nil
}

func ExampleOptionList() {
	l := Nil[Option[string]]()
	l = Cons(Some("foo"), l)
	l = Cons(Some("bar"), l)
	l = Cons(None[string](), l)

	head, tail, ok := l.Match().AsCons().Parameters()
	if ok {
		if head.Match().AsNone().OK() {
			fmt.Println("None")
		}
	}
	head, tail, ok = tail.Match().AsCons().Parameters()
	if ok {
		if v, ok := head.Match().AsSome().Parameters(); ok {
			fmt.Println("Some", v)
		}
	}
	head, tail, ok = tail.Match().AsCons().Parameters()
	if ok {
		if v, ok := head.Match().AsSome().Parameters(); ok {
			fmt.Println("Some", v)
		}
	}
	if tail.Match().AsNil().OK() {
		fmt.Println("Nil")
	}

	// Output:
	// None
	// Some bar
	// Some foo
	// Nil
}

func Nth[T any](l List[T], n int) Option[T] {
	var head T
	var tail List[T]
	var ok bool
	for i := 0; i <= n; i++ {
		head, tail, ok = l.Match().AsCons().Parameters()
		if !ok {
			return None[T]()
		}
		l = tail
	}
	return Some(head)
}

func ExampleNth() {
	// in Ino:
	//     l = Cons "foo" Nil
	//     opt = Nth l 0
	//     case Nth l 0 of
	//         Some v -> ...
	//       | None -> ...
	l := Cons("foo", Nil[string]())
	opt := Nth(l, 0)
	if v, ok := opt.Match().AsSome().Parameters(); ok {
		fmt.Println("Some", v)
	} else if opt.Match().AsNone().OK() {
		fmt.Println("None")
	}

	// Output:
	// Some foo
}