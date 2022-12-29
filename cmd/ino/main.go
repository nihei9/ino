package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nihei9/ino/code"
	"github.com/nihei9/ino/parser"
	"github.com/nihei9/ino/semantics"
)

var (
	packageName  string
	debugEnabled bool
)

func init() {
	flag.StringVar(&packageName, "package", "main", "pacakge name")
	flag.BoolVar(&debugEnabled, "debug", false, "if true, debug logging is enabled")
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()

	r, err := parser.Parse()
	if err != nil {
		fmt.Println("internal error:", err)
		return 1
	}
	if len(r.SynErrs) > 0 {
		for _, synErr := range r.SynErrs {
			fmt.Fprintln(os.Stderr, synErr)
		}
		return 1
	}

	var debugOut io.Writer
	if debugEnabled {
		f, err := os.OpenFile("ino.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		debugOut = f
	}
	a := &semantics.SemanticAnalyzer{
		DebugOut: debugOut,
	}
	err = a.Run(r.Tree)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	g := code.CodeGenerator{
		PkgName: packageName,
		Out:     os.Stdout,
	}
	err = g.Run(a.IR)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}
