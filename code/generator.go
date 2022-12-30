package code

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/nihei9/ino/ir"
)

var (
	//go:embed data_decl.tmpl
	dataDeclTmpl string

	//go:embed val_cons_decl.tmpl
	valConsDeclTmpl string
)

type CodeGenerator struct {
	PkgName string
	Out     io.Writer
}

func (g *CodeGenerator) Run(file *ir.File) error {
	f, err := genFile(file)
	if err != nil {
		return err
	}
	f.Name = ast.NewIdent(g.PkgName)
	fmt.Fprintln(g.Out, "// Code generated by ino. DO NOT EDIT.")
	err = format.Node(g.Out, token.NewFileSet(), f)
	if err != nil {
		return err
	}
	return nil
}

func genFile(file *ir.File) (*ast.File, error) {
	var decls []ast.Decl
	for _, decl := range file.Decls {
		d, err := genDecl(decl)
		if err != nil {
			return nil, err
		}
		decls = append(decls, d...)
	}
	return &ast.File{
		Decls: decls,
	}, nil
}

func genDecl(decl ir.Decl) ([]ast.Decl, error) {
	switch d := decl.(type) {
	case *ir.DataDecl:
		return genDataDecl(d)
	case *ir.ValConsDecl:
		return genValConsDecl(d)
	}
	return nil, fmt.Errorf("invalid declaration: %T", decl)
}

func genDataDecl(d *ir.DataDecl) ([]ast.Decl, error) {
	fm := template.FuncMap{
		// [T1 any, T2 any, ...]
		"genTyVarDecls": func() (string, error) {
			if d.TypeVarCount == 0 {
				return "", nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[T1 any")
			for i := 2; i <= d.TypeVarCount; i++ {
				fmt.Fprintf(&b, ", T%v any", i)
			}
			fmt.Fprintf(&b, "]")
			return b.String(), nil
		},
		// [T1, T2, ...]
		"genTyVarNames": func() (string, error) {
			if d.TypeVarCount == 0 {
				return "", nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[T1")
			for i := 2; i <= d.TypeVarCount; i++ {
				fmt.Fprintf(&b, ", T%v", i)
			}
			fmt.Fprintf(&b, "]")
			return b.String(), nil
		},
	}
	tmpl, err := template.New("dataDeclTmpl").Funcs(fm).Parse(dataDeclTmpl)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	err = tmpl.Execute(&b, map[string]string{
		"DataName":          d.Name,
		"MatcherStructName": "matcher_" + d.Name,
	})
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", b.String(), 0)
	if err != nil {
		panic(err)
	}
	return f.Decls, nil
}

func genValConsDecl(d *ir.ValConsDecl) ([]ast.Decl, error) {
	fm := template.FuncMap{
		// [T1 any, T2 any, ...]
		"genTyVarDecls": func() (string, error) {
			if d.TypeVarCount == 0 {
				return "", nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[T1 any")
			for i := 2; i <= d.TypeVarCount; i++ {
				fmt.Fprintf(&b, ", T%v any", i)
			}
			fmt.Fprintf(&b, "]")
			return b.String(), nil
		},
		// [T1, T2, ...]
		"genTyVarNames": func() (string, error) {
			if d.TypeVarCount == 0 {
				return "", nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[T1")
			for i := 2; i <= d.TypeVarCount; i++ {
				fmt.Fprintf(&b, ", T%v", i)
			}
			fmt.Fprintf(&b, "]")
			return b.String(), nil
		},
		"genFields": func() (string, error) {
			var b strings.Builder
			for i, p := range d.Params {
				ty, err := genType(p)
				if err != nil {
					return "", err
				}
				fmt.Fprintf(&b, "p%v %v\n", i+1, ty)
			}
			return b.String(), nil
		},
		"genParams": func() (string, error) {
			if len(d.Params) == 0 {
				return "", nil
			}

			var b strings.Builder
			ty, err := genType(d.Params[0])
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "p1 %v", ty)
			for i, p := range d.Params[1:] {
				ty, err := genType(p)
				if err != nil {
					return "", err
				}
				fmt.Fprintf(&b, ", p%v %v", i+2, ty)
			}
			return b.String(), nil
		},
		"genKeyValuePairs": func() (string, error) {
			var b strings.Builder
			for i := 1; i <= len(d.Params); i++ {
				fmt.Fprintf(&b, "p%v: p%v,\n", i, i)
			}
			return b.String(), nil
		},
		"genFieldsMethodResults": func() (string, error) {
			if len(d.Params) == 0 {
				return "", nil
			}

			var b strings.Builder
			ty, err := genType(d.Params[0])
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "p1 %v", ty)
			for i, p := range d.Params[1:] {
				ty, err := genType(p)
				if err != nil {
					return "", err
				}
				fmt.Fprintf(&b, ", p%v %v", i+2, ty)
			}
			fmt.Fprintf(&b, ", ok bool")
			return b.String(), nil
		},
		"genFieldsMethodReturn": func() (string, error) {
			if len(d.Params) == 0 {
				return "", nil
			}

			var b strings.Builder
			fmt.Fprintf(&b, "return x.x.p1")
			for i := 2; i <= len(d.Params); i++ {
				fmt.Fprintf(&b, ", x.x.p%v", i)
			}
			fmt.Fprintf(&b, ", true")
			return b.String(), nil
		},
	}
	tmpl, err := template.New("valConsDeclTmpl").Funcs(fm).Parse(valConsDeclTmpl)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	err = tmpl.Execute(&b, map[string]any{
		"DataName":           d.TyName,
		"TagName":            d.Name,
		"MatcherStructName":  "matcher_" + d.TyName,
		"TagStructName":      "tag_" + d.TyName + "_" + d.Name,
		"MaybeTagStructName": "maybeTag_" + d.TyName + "_" + d.Name,
		"HasFields":          len(d.Params) > 0,
	})
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", b.String(), 0)
	if err != nil {
		panic(err)
	}
	return f.Decls, nil
}

func genType(ty ir.Type) (ast.Expr, error) {
	switch t := ty.(type) {
	case *ir.BasicType:
		return ast.NewIdent(t.Name), nil
	case *ir.NamedType:
		return ast.NewIdent(t.Name), nil
	case *ir.TypeVar:
		return ast.NewIdent("T" + strconv.Itoa(t.Num)), nil
	}
	return nil, fmt.Errorf("invalid type: %T", ty)
}
