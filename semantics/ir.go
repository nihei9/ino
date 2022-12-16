package semantics

import (
	"fmt"
	"strconv"

	"github.com/nihei9/ino/ir"
	"github.com/nihei9/ino/parser"
)

type irBuilder struct {
	tyEnv  *tyEnv
	valEnv *valEnv
}

func (b *irBuilder) enter(sym symbol) error {
	ee, ok := b.valEnv.lookup(sym)
	if !ok {
		return fmt.Errorf("failed to enter an environment of %v", sym)
	}
	b.valEnv = &valEnv{ee.valEnv.bindingsTree}
	b.tyEnv = &tyEnv{ee.tyEnv.bindingsTree}
	return nil
}

func (b *irBuilder) leave() {
	if b.valEnv.parent == nil || b.tyEnv.parent == nil {
		panic(fmt.Errorf("cannot leave because parents is nil"))
	}
	b.valEnv = &valEnv{b.valEnv.parent}
	b.tyEnv = &tyEnv{b.tyEnv.parent}
}

func (b *irBuilder) run(root *parser.Node) (*ir.File, error) {
	return b.root(root)
}

func (b *irBuilder) root(root *parser.Node) (*ir.File, error) {
	decls, err := b.decls(root.Children[0])
	if err != nil {
		return nil, err
	}
	return &ir.File{
		Decls: decls,
	}, nil
}

func (b *irBuilder) decls(node *parser.Node) ([]ir.Decl, error) {
	decls := make([]ir.Decl, 0, len(node.Children))
	for _, c := range node.Children {
		ds, err := b.decl(c)
		if err != nil {
			return nil, err
		}
		decls = append(decls, ds...)
	}
	return decls, nil
}

func (b *irBuilder) decl(node *parser.Node) ([]ir.Decl, error) {
	d := node.Children[0]
	switch d.KindName {
	case "const":
		d, err := b.constVal(d)
		if err != nil {
			return nil, err
		}
		return []ir.Decl{d}, nil
	case "func":
		d, err := b.fun(d)
		if err != nil {
			return nil, err
		}
		return []ir.Decl{d}, nil
	case "data":
		return b.data(d)
	}
	return nil, fmt.Errorf("invalid declaration node kind: %v", d.KindName)
}

func (b *irBuilder) constVal(node *parser.Node) (ir.Decl, error) {
	constName := node.Children[0].Text
	var result ir.Type
	{
		ee, ok := b.valEnv.lookup(symbol(constName))
		if !ok {
			return nil, fmt.Errorf("constant value definition not found: %v", constName)
		}
		var err error
		result, err = genType(ee.ty)
		if err != nil {
			return nil, fmt.Errorf("failed to generate a result type of a constant value %v: %w", constName, err)
		}
	}
	body, err := b.expr(node.Children[2])
	if err != nil {
		return nil, err
	}
	return &ir.FuncDecl{
		Name:   constName,
		Result: result,
		Body:   body,
	}, nil
}

func (b *irBuilder) fun(node *parser.Node) (ir.Decl, error) {
	funcName := node.Children[0].Text

	var funcTy *funcType
	{
		ee, ok := b.valEnv.lookup(symbol(funcName))
		if !ok {
			return nil, fmt.Errorf("function definition not found: %v", funcName)
		}
		funcTy, ok = ee.ty.(*funcType)
		if !ok {
			return nil, fmt.Errorf("symbol is not a function: %v: %T", funcName, ee.ty)
		}
	}

	var params []*ir.Param
	{
		ps := node.Children[1]
		params = make([]*ir.Param, len(ps.Children))
		for i, p := range ps.Children {
			pName := p.Children[0].Text
			pTy, err := genType(funcTy.params[i])
			if err != nil {
				return nil, fmt.Errorf("failed to generate a parameter type of a function %v: %w", funcName, err)
			}
			params[i] = &ir.Param{
				Name: pName,
				Ty:   pTy,
			}
		}
	}

	var result ir.Type
	{
		var err error
		result, err = genType(funcTy.result)
		if err != nil {
			return nil, fmt.Errorf("failed to generate a result type of a function %v: %w", funcName, err)
		}
	}

	b.enter(symbol(funcName))
	defer b.leave()
	body, err := b.expr(node.Children[3])
	if err != nil {
		return nil, err
	}

	return &ir.FuncDecl{
		Name:   funcName,
		Params: params,
		Result: result,
		Body:   body,
	}, nil
}

func (b *irBuilder) expr(node *parser.Node) (ir.Expr, error) {
	switch node.Children[0].KindName {
	case "call":
		call := node.Children[0]
		fun, err := b.expr(call)
		if err != nil {
			return nil, err
		}
		args := make([]ir.Expr, len(call.Children[1].Children))
		for i, arg := range call.Children[1].Children {
			a, err := b.expr(arg.Children[0])
			if err != nil {
				return nil, err
			}
			args[i] = a
		}
		return &ir.CallExpr{
			Func: fun,
			Args: args,
		}, nil
	case "id":
		id := node.Children[0].Text
		ee, ok := b.valEnv.lookup(symbol(id))
		if !ok {
			return nil, fmt.Errorf("symbol is not defined: %v", id)
		}
		return &ir.IdentExpr{
			Ident:    id,
			Constant: ee.constant,
		}, nil
	case "int":
		v, err := strconv.Atoi(node.Children[0].Text)
		if err != nil {
			return nil, fmt.Errorf("failed to parse an integer literal: %w", err)
		}
		return &ir.IntLit{
			Val: v,
		}, nil
	case "string":
		return &ir.StringLit{
			Val: node.Children[0].Text,
		}, nil
	case "add":
		lhs, err := b.expr(node.Children[1])
		if err != nil {
			return nil, err
		}
		rhs, err := b.expr(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &ir.BinaryExpr{
			LHS: lhs,
			Op:  ir.OpAdd,
			RHS: rhs,
		}, nil
	case "sub":
		lhs, err := b.expr(node.Children[1])
		if err != nil {
			return nil, err
		}
		rhs, err := b.expr(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &ir.BinaryExpr{
			LHS: lhs,
			Op:  ir.OpSub,
			RHS: rhs,
		}, nil
	case "mul":
		lhs, err := b.expr(node.Children[1])
		if err != nil {
			return nil, err
		}
		rhs, err := b.expr(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &ir.BinaryExpr{
			LHS: lhs,
			Op:  ir.OpMul,
			RHS: rhs,
		}, nil
	case "div":
		lhs, err := b.expr(node.Children[1])
		if err != nil {
			return nil, err
		}
		rhs, err := b.expr(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &ir.BinaryExpr{
			LHS: lhs,
			Op:  ir.OpDiv,
			RHS: rhs,
		}, nil
	case "mod":
		lhs, err := b.expr(node.Children[1])
		if err != nil {
			return nil, err
		}
		rhs, err := b.expr(node.Children[2])
		if err != nil {
			return nil, err
		}
		return &ir.BinaryExpr{
			LHS: lhs,
			Op:  ir.OpMod,
			RHS: rhs,
		}, nil
	}
	return nil, fmt.Errorf("invalid expression: %T", node.Children[0].KindName)
}

func (b *irBuilder) data(node *parser.Node) ([]ir.Decl, error) {
	dataName := node.Children[0].Text
	conss := node.Children[1]
	decls := make([]ir.Decl, 0, 1+len(conss.Children))
	decls = append(decls, &ir.DataDecl{
		Name: dataName,
	})
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
			param, err := genType(t)
			if err != nil {
				return nil, fmt.Errorf("failed to generate a type of a value constructor of %v: %w", dataName, err)
			}
			params[i] = param
		}
		decls = append(decls, &ir.ValConsDecl{
			Name:   consName,
			TyName: dataName,
			Params: params,
		})
	}
	return decls, nil
}

func genType(ty declType) (ir.Type, error) {
	switch t := ty.(type) {
	case basicType:
		return &ir.BasicType{
			Name: t.String(),
		}, nil
	case *dataType:
		return &ir.NamedType{
			Name: string(t.name),
		}, nil
	}
	return nil, fmt.Errorf("invalid type: %T", ty)
}
