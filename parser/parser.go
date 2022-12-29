//go:generate vartan compile ino.vartan -o ino.json
//go:generate vartan-go ino.json --package parser
package parser

import (
	"fmt"
	"os"
	"strings"
)

type ParseResult struct {
	Tree    *Node
	SynErrs []string
}

func Parse() (*ParseResult, error) {
	toks, err := NewTokenStream(os.Stdin)
	if err != nil {
		return nil, err
	}
	gram := NewGrammar()
	tb := NewDefaultSyntaxTreeBuilder()
	p, err := NewParser(toks, gram, SemanticAction(NewASTActionSet(gram, tb)))
	if err != nil {
		return nil, err
	}
	err = p.Parse()
	if err != nil {
		return nil, err
	}
	var synErrs []string
	for _, e := range p.SyntaxErrors() {
		synErrs = append(synErrs, genErrorMessage(e, gram))
	}
	if len(synErrs) > 0 {
		return &ParseResult{
			SynErrs: synErrs,
		}, nil
	}
	return &ParseResult{
		Tree: tb.Tree(),
	}, nil
}

func genErrorMessage(synErr *SyntaxError, gram Grammar) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v:%v: %v: ", synErr.Row+1, synErr.Col+1, synErr.Message)
	tok := synErr.Token
	switch {
	case tok.EOF():
		fmt.Fprintf(&b, "<eof>")
	case tok.Invalid():
		fmt.Fprintf(&b, "'%v' (<invalid>)", string(tok.Lexeme()))
	default:
		if term := gram.Terminal(tok.TerminalID()); term != "" {
			fmt.Fprintf(&b, "'%v' (%v)", string(tok.Lexeme()), term)
		} else {
			fmt.Fprintf(&b, "'%v'", string(tok.Lexeme()))
		}
	}
	if len(synErr.ExpectedTerminals) > 0 {
		fmt.Fprintf(&b, ": expected: %v", synErr.ExpectedTerminals[0])
		for _, t := range synErr.ExpectedTerminals[1:] {
			fmt.Fprintf(&b, ", %v", t)
		}
	}
	return b.String()
}
