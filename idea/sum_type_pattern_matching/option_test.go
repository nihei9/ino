package main

import "fmt"

func ExampleOption() {
	{
		s1 := Some(Some(I(100)))
		s2 := Some(Some(I(100)))
		if s1.Eq(s2) && s2.Eq(s1) {
			fmt.Println("s1 == s2")
		}
	}

	{
		s1 := Some(Some(None[I]()))
		s2 := Some(Some(I(100)))
		if !s1.Eq(s2) && !s2.Eq(s1) {
			fmt.Println("s1 != s2")
		}
	}

	{
		result, err := Match(Some(I(100)),
			CaseSome(Some(I(100)), func(p1 I) string {
				return "Some(100)"
			}),
			CaseNone(None[I](), func() string {
				return "None"
			}),
		)
		if err != nil {
			fmt.Print(err)
		} else {
			fmt.Println(result)
		}
	}

	// Output:
	// s1 == s2
	// s1 != s2
	// Some(100)
}
