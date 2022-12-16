# Ino

Ino is a tiny static-typed language and compiles source code into golang.

## Ino syntax and available Go APIs

### `data` type

#### Example

Ino (source code):

```
data Pen = BallpointPen string
         | FountainPen
```

Go (generated code):

The following three public APIs are available.

* `Pen` interface
* `BallpointPen` function that returns a Pen object
* `FountainPen` function that returns a Pen object

```go
bB := BallpointPen("black")
ok := bB.Maybe().BallpointPen().OK() // true
ok = bB.Maybe().FountainPen().OK()   // false

f := FountainPen()
ok = f.Maybe().FountainPen().OK()  // true
ok = f.Maybe().BallpointPen().OK() // false
```
