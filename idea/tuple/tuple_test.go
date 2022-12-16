package main

import "fmt"

// in Ino:
//     type Kanji = (int, string, string)

func Kanji(d int, e string, j string) *tuple3[int, string, string] {
	return tuple3_new(d, e, j)
}

func ExampleKanji() {
	k := Kanji(5, "five", "五")
	d, e, j := k.Elements()
	fmt.Println(d, e, j)

	// Output:
	// 5 five 五
}
