package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nihei9/ino/code"
	"github.com/nihei9/ino/multierr"
	"github.com/nihei9/ino/parser"
	"github.com/nihei9/ino/semantics"
)

var flags = struct {
	packageName  string
	debugEnabled bool
}{}

func init() {
	flag.StringVar(&flags.packageName, "package", "main", "pacakge name")
	flag.BoolVar(&flags.debugEnabled, "debug", false, "if true, debug logging is enabled")
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()

	files, err := findInoFiles()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	for _, f := range files {
		wg, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = compile(f, wg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	return 0
}

func findInoFiles() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return filepath.Glob(filepath.Join(wd, "*.ino"))
}

func compile(srcPath string, outDir string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := parser.Parse(f)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	if len(r.SynErrs) > 0 {
		var errs []error
		for _, synErr := range r.SynErrs {
			errs = append(errs, fmt.Errorf(synErr))
		}
		return multierr.Bundle(errs...)
	}

	var debugOut io.Writer
	if flags.debugEnabled {
		f, err := os.OpenFile("ino.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		debugOut = f
	}
	a := &semantics.SemanticAnalyzer{
		DebugOut: debugOut,
	}
	err = a.Run(r.Tree)
	if err != nil {
		return err
	}

	srcFileName := filepath.Base(srcPath)
	outFileName := strings.TrimSuffix(srcFileName, ".ino") + ".go"
	g := code.CodeGenerator{
		PkgName:     flags.packageName,
		OutDir:      outDir,
		OutFileName: outFileName,
	}
	err = g.Run(a.IR)
	if err != nil {
		return err
	}

	return nil
}
