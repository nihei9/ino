//go:generate vartan compile ino.vartan -o ino.json
//go:generate vartan-go ino.json --package parser
package parser

import (
	"fmt"
	"os"
)

type ParseResult struct {
	Tree    *Node
	SynErrs []*SyntaxError
	Grammar Grammar
}

func Parse() (*ParseResult, error) {
	toks, err := NewTokenStream(os.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gram := NewGrammar()
	tb := NewDefaultSyntaxTreeBuilder()
	p, err := NewParser(toks, gram, SemanticAction(NewASTActionSet(gram, tb)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = p.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	synErrs := p.SyntaxErrors()
	if len(synErrs) > 0 {
		return &ParseResult{
			SynErrs: synErrs,
			Grammar: gram,
		}, nil
	}
	return &ParseResult{
		Tree: tb.Tree(),
	}, nil
}
