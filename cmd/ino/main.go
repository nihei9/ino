package main

import (
	"fmt"
	"io"
	"os"

	"github.com/nihei9/ino/code"
	"github.com/nihei9/ino/parser"
	"github.com/nihei9/ino/semantics"
)

func main() {
	r, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(r.SynErrs) > 0 {
		for _, synErr := range r.SynErrs {
			printSyntaxError(os.Stderr, synErr, r.Grammar)
		}
		os.Exit(1)
	}

	a := &semantics.SemanticAnalyzer{
		DebugOut: os.Stderr,
	}
	err = a.Run(r.Tree)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	g := code.CodeGenerator{
		PkgName: "main",
		Out: os.Stdout,
	}
	err = g.Run(a.IR)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printSyntaxError(w io.Writer, synErr *parser.SyntaxError, gram parser.Grammar) {
	var msg string
	tok := synErr.Token
	switch {
	case tok.EOF():
		msg = "<eof>"
	case tok.Invalid():
		msg = fmt.Sprintf("'%v' (<invalid>)", string(tok.Lexeme()))
	default:
		if term := gram.Terminal(tok.TerminalID()); term != "" {
			msg = fmt.Sprintf("'%v' (%v)", string(tok.Lexeme()), term)
		} else {
			msg = fmt.Sprintf("'%v'", string(tok.Lexeme()))
		}
	}
	fmt.Fprintf(w, "%v:%v: %v: %v", synErr.Row+1, synErr.Col+1, synErr.Message, msg)

	if len(synErr.ExpectedTerminals) > 0 {
		fmt.Fprintf(w, ": expected: %v", synErr.ExpectedTerminals[0])
		for _, t := range synErr.ExpectedTerminals[1:] {
			fmt.Fprintf(w, ", %v", t)
		}
	}

	fmt.Fprintf(w, "\n")
}
