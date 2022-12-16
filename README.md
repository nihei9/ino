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

### `data` type

#### Example

Ino:

```
data Pen
    = BallpointPen string
    | FountainPen
    ;
```

Go:

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
