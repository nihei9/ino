package semantics

import (
	"fmt"

	"github.com/nihei9/ino/ir"
	"github.com/nihei9/ino/parser"
)

type irBuilder struct {
	tyEnv  *tyEnv
	valEnv *valEnv
}

func (b *irBuilder) run(root *parser.Node) (*ir.File, error) {
	return b.buildRoot(root)
}

func (b *irBuilder) buildRoot(root *parser.Node) (*ir.File, error) {
	decls, err := b.buildDecls(root.Children[0])
	if err != nil {
		return nil, err
	}
	return &ir.File{
		Decls: decls,
	}, nil
}

func (b *irBuilder) buildDecls(node *parser.Node) ([]ir.Decl, error) {
	decls := make([]ir.Decl, 0, len(node.Children))
	for _, c := range node.Children {
		ds, err := b.buildDecl(c)
		if err != nil {
			return nil, err
		}
		decls = append(decls, ds...)
	}
	return decls, nil
}

func (b *irBuilder) buildDecl(node *parser.Node) ([]ir.Decl, error) {
	d := node.Children[0]
	switch d.KindName {
	case "data":
		return b.buildData(d)
	}
	return nil, fmt.Errorf("invalid declaration node kind: %v", d.KindName)
}

func (b *irBuilder) buildData(node *parser.Node) ([]ir.Decl, error) {
	dataName := node.Children[0].Text
	tyVars := node.Children[1]
	conss := node.Children[2]
	decls := make([]ir.Decl, 0, 1+len(conss.Children))
	decls = append(decls, &ir.DataDecl{
		Name:         dataName,
		TypeVarCount: len(tyVars.Children),
	})
	tyVar2Num := map[symbol]int{}
	for i, v := range tyVars.Children {
		tyVar2Num[symbol(v.Children[0].Text)] = i + 1
	}
	for _, cons := range conss.Children {
		consName := cons.Children[0].Text
		ee, ok := b.valEnv.lookup(symbol(consName))
		if !ok {
			return nil, fmt.Errorf("symbol not found: %v", consName)
		}
		consTy, ok := ee.ty.(*funcType)
		if !ok {
			return nil, fmt.Errorf("value constructor must be a function type: %T", ee.ty)
		}
		params := make([]ir.Type, len(consTy.params))
		for i, t := range consTy.params {
			param, err := genType(t, tyVar2Num)
			if err != nil {
				return nil, fmt.Errorf("failed to generate a type of a value constructor %v of %v: %w", consName, dataName, err)
			}
			params[i] = param
		}
		decls = append(decls, &ir.ValConsDecl{
			Name:         consName,
			TyName:       dataName,
			Params:       params,
			TypeVarCount: len(tyVars.Children),
		})
	}
	return decls, nil
}

func genType(ty declType, tyVar2Num map[symbol]int) (ir.Type, error) {
	switch t := ty.(type) {
	case basicType:
		return &ir.BasicType{
			Name: t.String(),
		}, nil
	case *dataType:
		return &ir.NamedType{
			Name: string(t.name),
		}, nil
	case *concreteType:
		abTy, err := genType(t.abstractTy, tyVar2Num)
		if err != nil {
			return nil, err
		}
		args := make([]ir.Type, len(t.args))
		for i, p := range t.args {
			pTy, err := genType(p, tyVar2Num)
			if err != nil {
				return nil, err
			}
			args[i] = pTy
		}
		return &ir.ConcreteType{
			AbstractTy: abTy,
			Args:       args,
		}, nil
	case *typeVar:
		return &ir.TypeVar{
			Num: tyVar2Num[t.name],
		}, nil
	}
	return nil, fmt.Errorf("invalid type: %T", ty)
}
