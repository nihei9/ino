package main

import "fmt"

func ExampleOption() {
	optV := Some("foo")
	none := None[string]()

	if v, ok := optV.Match().AsSome().Parameters(); ok {
		fmt.Println(v)
	}
	if none.Match().AsNone().OK() {
		fmt.Print("None")
	}

	// Output:
	// foo
	// None
}
