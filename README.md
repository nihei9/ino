# Ino

[![Test](https://github.com/nihei9/ino/actions/workflows/test.yml/badge.svg)](https://github.com/nihei9/ino/actions/workflows/test.yml)

Ino is a DSL to generate data structures for golang.

## Install

```
go install github.com/nihei9/ino/cmd/ino@latest
```

## Usage

First, save your source code written in Ino to a file with the extension `.ino`. In this example, let's say you saved the file named `example.ino`.

Next, convert the source code in Ino to golang using `ino` command. `--package` flag allows you to change the package name. The default package name is `main`.

```
ino --package example
```

Then you will get a file named `example.go`.

You can also use `go:generate` directive.

```
//go:generate ino --package example
```

## Ino syntax and available Go APIs

### `data`

`data` keyword allows you to generate a [sum type](https://en.wikipedia.org/wiki/Tagged_union).

#### Example: Basic usage

##### In Ino

```
data Pen
    = BallpointPen string
    | FountainPen
    ;
```

##### Go APIs generated

The following three public APIs are available.

* `type Pen interface`
* `func BallpointPen(string) Pen`: returns a `BallpointPen` variant of `Pen` type
* `func FountainPen() Pen`: returns a `FountainPen` variant of `Pen` type

Now, you can check that a value of `Pen` is which variant using `OK` method and can retrieve fields using `Fields` method.
`Fields` method isn't available in `FountainPen` because it has no fields.

```go
bpB := BallpointPen("Black")
ok := bpB.Maybe().BallpointPen().OK()            // true
ok = bpB.Maybe().FountainPen().OK()              // false
color, ok := bpB.Maybe().BallpointPen().Fields() // "Black", true
color, ok = bpB.Maybe().FountainPen().Fields()   // "", false

fp := FountainPen()
ok = fp.Maybe().FountainPen().OK()  // true
ok = fp.Maybe().BallpointPen().OK() // false
```

The following helpful APIs are also available.

* `func ApplyToBallpointPen[T any](x Pen, f func(string) T) (T, bool)`: when `x` is a `BallpointPen` variant, applies `f` to `x`'s fields and retuns the result and `true`.
* `func ApplyToFountainPen[T any](x Pen, f func() T) (T, bool)`: when `x` is a `FountainPen` variant, runs `f` and returns the result and `true`

The result type of `ApplyTo*` function is polymorphic, so you can use an arbitrary type for the result type of the callback function `f`.

```go
color2Num := func(color string) int {
    switch color {
    case "Black": return 1
    default: return 9
    }
}

bpB := BallpointPen("Black")
bpR := BallpointPen("Red")
fp := FountainPen()

v, ok := ApplyToBallpointPen(bpB, color2Num) // 1, true
v, ok := ApplyToBallpointPen(bpR, color2Num) // 9, true
v, ok = ApplyToBallpointPen(fp, color2Num)   // 0, false
```

#### Example: Generics #1

You can define polymorphic types using type variables.

##### In Ino

Type variables can follow a data type name. In this example, `a` is a type variable.

```
data Option a
    = None
    | Some a
    ;
```

##### Go APIs generatd

* `type Option[T any] interface`
* `func None[T any]() Option[T]`: returns a `None` variant of `Option` type
* `func Some[T any](p1 T1) Option[T]`: returns a `Some` variant of `Option` type

```go
s1 := Some(100)           // Option[int]
n := None[int]            // Option[int]
s2 := Some(Some("Hello")) // Option[Option[string]]

num, ok := s1.Maybe().Some().Fields()  // 100, true
opt, ok := s2.Maybe().Some().Fields()  // Some("Hello"), true
str, ok := opt.Maybe().Some().Fields() // "Hello", true
```

`ApplyTo*` functions are also available.

* `func ApplyToNone[T any, U any](x Option[T], f func() U) (U, bool)`
* `func ApplyToSome[T any, U any](x Option[T], f func(T) U) (U, bool)`

#### Example: Generics #2

Polymorphic-type literals must be enclosed in parentheses.

```
data List a
    = Nil
    | Cons a (List a)
    ;
```
