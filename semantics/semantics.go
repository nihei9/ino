package semantics

import (
	"fmt"
	"io"
	"os"

	"github.com/nihei9/ino/ir"
	"github.com/nihei9/ino/parser"
)

type SemanticAnalyzer struct {
	DebugOut io.Writer
	IR       *ir.File
}

func (a *SemanticAnalyzer) Run(root *parser.Node) error {
	if a.DebugOut != nil {
		fmt.Fprintln(a.DebugOut, "=================")
		fmt.Fprintln(a.DebugOut, "Semantic Analysis")
		fmt.Fprintln(a.DebugOut, "=================")
	}
	eb := &environmentBuilder{}
	err := eb.run(root)
	if err != nil {
		if a.DebugOut != nil {
			fmt.Fprintln(a.DebugOut, "Environment building ... Failed")
			fmt.Fprintln(a.DebugOut, "Errors:")
			for _, e := range eb.errs {
				fmt.Fprintln(a.DebugOut, e)
			}
		}
		// FIXME
		for _, e := range eb.errs {
			fmt.Println(e)
		}
		return err
	}
	if a.DebugOut != nil {
		fmt.Fprintln(a.DebugOut, "Environment building ... Passed")
		fmt.Fprintln(a.DebugOut, "Type Environment:")
		eb.tyEnvRoot.print(os.Stdout)
		fmt.Fprintln(a.DebugOut, "--------")
		fmt.Fprintln(a.DebugOut, "Value Environment:")
		eb.valEnvRoot.print(os.Stdout)
		fmt.Fprintln(a.DebugOut, "--------")
	}

	sa := &typeChecker{
		ast:    root,
		tyEnv:  eb.tyEnvRoot,
		valEnv: eb.valEnvRoot,
	}
	err = sa.run()
	if err != nil {
		if a.DebugOut != nil {
			fmt.Fprintln(a.DebugOut, "Type checking ... Failed")
		}
		return err
	}
	if a.DebugOut != nil {
		fmt.Fprintln(a.DebugOut, "Type checking ... Passed")
		fmt.Fprintln(a.DebugOut, "Type Environment:")
		eb.tyEnvRoot.print(os.Stdout)
		fmt.Fprintln(a.DebugOut, "--------")
		fmt.Fprintln(a.DebugOut, "Value Environment:")
		eb.valEnvRoot.print(os.Stdout)
		fmt.Fprintln(a.DebugOut, "--------")
	}

	irb := &irBuilder{
		tyEnv:  &tyEnv{eb.tyEnvRoot.child},
		valEnv: &valEnv{eb.valEnv.child},
	}
	ir, err := irb.run(root)
	if err != nil {
		if a.DebugOut != nil {
			fmt.Fprintln(a.DebugOut, "IR generating ... Failed")
		}
		return err
	}
	a.IR = ir
	if a.DebugOut != nil {
		fmt.Fprintln(a.DebugOut, "IR generating ... Passed")
	}

	return nil
}
