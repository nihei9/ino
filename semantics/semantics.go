package semantics

import (
	"fmt"
	"io"
	"strings"

	"github.com/nihei9/ino/ir"
	"github.com/nihei9/ino/multierr"
	"github.com/nihei9/ino/parser"
)

type logger struct {
	w io.Writer
}

func (l *logger) header1(msg string) {
	if l.w == nil {
		return
	}
	fence := strings.Repeat("=", len(msg))
	fmt.Fprintln(l.w, fence)
	fmt.Fprintln(l.w, msg)
	fmt.Fprintln(l.w, fence)
}

func (l *logger) header2(msg string) {
	if l.w == nil {
		return
	}
	fence := strings.Repeat("-", len(msg))
	fmt.Fprintln(l.w, fence)
	fmt.Fprintln(l.w, msg)
	fmt.Fprintln(l.w, fence)
}

func (l *logger) write(format string, a ...any) {
	if l.w == nil {
		return
	}
	fmt.Fprintf(l.w, format+"\n", a...)
}

type SemanticAnalyzer struct {
	DebugOut io.Writer
	IR       *ir.File
}

func (a *SemanticAnalyzer) Run(root *parser.Node) error {
	l := &logger{
		w: a.DebugOut,
	}

	l.header1("Semantic Analysis")

	l.header2("Environment building")
	eb := &environmentBuilder{}
	err := eb.run(root)
	if err != nil {
		merr := multierr.Bundle(eb.errs...)
		l.write("Failed")
		l.write(err.Error())
		l.write(merr.Error())
		return merr
	}
	l.write("Passed")
	l.write("Type Environment:")
	{
		var b strings.Builder
		eb.tyEnvRoot.print(&b)
		l.write(b.String())
	}
	l.write("--------")
	l.write("Value Environment:")
	{
		var b strings.Builder
		eb.valEnvRoot.print(&b)
		l.write(b.String())
	}

	l.header2("Type checking")
	sa := &typeChecker{
		ast:    root,
		tyEnv:  eb.tyEnvRoot,
		valEnv: eb.valEnvRoot,
	}
	err = sa.run()
	if err != nil {
		l.write("Failed")
		l.write(err.Error())
		return err
	}
	l.write("Passed")
	l.write("Type Environment:")
	{
		var b strings.Builder
		eb.tyEnvRoot.print(&b)
		l.write(b.String())
	}
	l.write("--------")
	l.write("Value Environment:")
	{
		var b strings.Builder
		eb.valEnvRoot.print(&b)
		l.write(b.String())
	}

	l.header2("IR building")
	irb := &irBuilder{
		tyEnv:  &tyEnv{eb.tyEnvRoot.child},
		valEnv: &valEnv{eb.valEnv.child},
	}
	ir, err := irb.run(root)
	if err != nil {
		l.write("Failed")
		l.write(err.Error())
		return err
	}
	a.IR = ir
	l.write("Passed")

	return nil
}
