package main

import "fmt"

func ExamplePen() {
	bBlackMid := BallpointPen("black", 0.8)
	f := FountainPen()

	if bBlackMid.Match().AsBallpointPen().OK() {
		fmt.Println("This is a ball-point pen.")
	}

	if !bBlackMid.Match().AsFountainPen().OK() {
		fmt.Println("This is not a fountain pen.")
	}

	if color, size, ok := bBlackMid.Match().AsBallpointPen().Parameters(); ok {
		fmt.Println(color, size)
	}

	fmt.Println("---")

	match(bBlackMid)
	match(f)

	// Output:
	// This is a ball-point pen.
	// This is not a fountain pen.
	// black 0.8
	// ---
	// ball-point pen black 0.8
	// fountain pen
}

func match(p Pen) {
	if color, size, ok := p.Match().AsBallpointPen().Parameters(); ok {
		fmt.Println("ball-point pen", color, size)
	} else if ok := p.Match().AsFountainPen().OK(); ok {
		fmt.Println("fountain pen")
	}
}
