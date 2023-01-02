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

Then you will get files `example.go` and `ino_builtin.go`.

You can also use `go:generate` directive.

```
//go:generate ino --package example
```

## Ino syntax and available Go APIs

### Encoding

Source code is Unicode text encoded in UTF-8.

### Lexical elements

#### Identifiers

Identifiers name types and variables. The form of identifiers is `[A-Za-z][A-Za-z0-9]*` in the regular expression. Identifiers in Ino are limited in the number of characters we can use compared to golang.

#### Keywords

The following keywords may not be used as identifiers.

* `data`

### Built-in types

Ino currently supports `Int` and `String` built-in types, corresponding to `int` and `string` of golang, respectively.

### `data` type

`data` keyword allows you to generate a [sum type](https://en.wikipedia.org/wiki/Tagged_union).

#### Example: Basic usage

##### In Ino

```
data Pen
    = BallpointPen String
    | FountainPen
    ;
```

##### Go APIs generated

The following three public APIs are available.

* `type Pen interface`
* `func BallpointPen(String) Pen`: returns a `BallpointPen` variant of `Pen` type
* `func FountainPen() Pen`: returns a `FountainPen` variant of `Pen` type

Now, you can check that a value of `Pen` is which variant using `OK` method and can retrieve fields using `Fields` method.
`Fields` method isn't available in `FountainPen` because it has no fields.

```go
bpB := BallpointPen(String("Black"))
ok := bpB.Maybe().BallpointPen().OK()            // true
ok = bpB.Maybe().FountainPen().OK()              // false
color, ok := bpB.Maybe().BallpointPen().Fields() // String("Black"), true
color, ok = bpB.Maybe().FountainPen().Fields()   // String(""), false

fp := FountainPen()
ok = fp.Maybe().FountainPen().OK()  // true
ok = fp.Maybe().BallpointPen().OK() // false
```

The following helpful APIs are also available.

* `func NewPenCaseSet[U any](...*case_Pen[U]) (*PenCaseSet[U], error)` and `func (s *PenCaseSet[U]) Match(x Pen) (U, error)`: run pattern matching.
* `func ApplyToBallpointPen[T any](x Pen, f func(String) T) (T, bool)`: when `x` is a `BallpointPen` variant, applies `f` to `x`'s fields and retuns the result and `true`.
* `func ApplyToFountainPen[T any](x Pen, f func() T) (T, bool)`: when `x` is a `FountainPen` variant, runs `f` and returns the result and `true`.

The result type of `Match` method and `ApplyTo*` function are polymorphic, so you can use an arbitrary type for the result type of the callback function.

Pattern matching:

:warning: `New*CaseSet` function performs exhaustiveness checking, but the checking guarantees only the exhaustiveness of tags. In other words, the exhaustive of parameters is not guaranteed.

```go
cases, err := NewPenCaseSet(
    CaseBallpointPen(BallpointPen(String("Black")), func(color String) string {
        return "ballpoint-"+string(color)
    }),
    CaseFountainPen(FountainPen(), func() string {
        return "fountain"
    }),
)
if err != nil {
    // When the above cases are not exhaustive pattern, an error occurs.
}

pen := BallpointPen(String("Black"))

result, err := cases.Match(pen)
if err != nil {
    // When `pen` doesn't match any of the patterns, an error occurs.
}
fmt.Print(result) // ballpoint-Black
```

`ApplyTo*`:

```go
color2Num := func(color String) int {
    switch color {
    case "Black": return 1
    default: return 9
    }
}

bpB := BallpointPen(String("Black"))
bpR := BallpointPen(String("Red"))
fp := FountainPen()

v, ok := ApplyToBallpointPen(bpB, color2Num) // 1, true
v, ok = ApplyToBallpointPen(bpR, color2Num)  // 9, true
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

* `type Option[T Eqer] interface`
* `func None[T Eqer]() Option[T]`: returns a `None` variant of `Option` type
* `func Some[T Eqer](p1 T) Option[T]`: returns a `Some` variant of `Option` type

```go
s1 := Some(Int(100))              // Option[Int]
n := None[Int]()                  // Option[Int]
s2 := Some(Some(String("Hello"))) // Option[Option[String]]

num, ok := s1.Maybe().Some().Fields()  // Int(100), true
opt, ok := s2.Maybe().Some().Fields()  // Some(String("Hello")), true
str, ok := opt.Maybe().Some().Fields() // String("Hello"), true
```

Pattern matching APIs and `ApplyTo*` functions are also available.

* `func NewOptionCaseSet[T Eqer, U any](...*case_Option[T, U]) (*OptionCaseSet[T, U], error)` and `func (s *OptionCaseSet[T, U]) Match(x Option[T]) (U, error)`
* `func ApplyToNone[T Eqer, U any](x Option[T], f func() U) (U, bool)`
* `func ApplyToSome[T Eqer, U any](x Option[T], f func(T) U) (U, bool)`

#### Example: Generics #2

Polymorphic-type literals must be enclosed in parentheses.

```
data List a
    = Nil
    | Cons a (List a)
    ;
```
