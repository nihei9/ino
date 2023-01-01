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

* `Pen` interface
* `BallpointPen` function (constructor) that returns a `Pen` object
* `FountainPen` function (constructor) that returns a `Pen` object

Now, you can check that a value of `Pen` is which variant using `OK` method and can retrieve fields using `Fields` method.

```go
bB := BallpointPen("Black")
ok := bB.Maybe().BallpointPen().OK()            // true
ok = bB.Maybe().FountainPen().OK()              // false
color, ok := bB.Maybe().BallpointPen().Fields() // "Black", true
color, ok = bB.Maybe().FountainPen().Fields()   // "", false

f := FountainPen()
ok = f.Maybe().FountainPen().OK()  // true
ok = f.Maybe().BallpointPen().OK() // false
```

:eyes: `Fields` method isn't available in `FountainPen` because it has no fields.

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

* `Option[T1 any]` interface
* `None` function that returns a `Option` object
* `Some[T1 any]` function that returns a `Option` object

```go
s1 := Some(100)           // Option[int]
n := None[int]            // Option[int]
s2 := Some(Some("Hello")) // Option[Option[string]]

num, ok := s1.Maybe().Some().Fields()  // 100, true
opt, ok := s2.Maybe().Some().Fields()  // Some("Hello"), true
str, ok := opt.Maybe().Some().Fields() // "Hello", true
```

#### Example: Generics #2

Polymorphic-type literals must be enclosed in parentheses.

```
data List a
    = Nil
    | Cons a (List a)
    ;
```
