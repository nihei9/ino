package semantics

import (
	"fmt"

	"github.com/nihei9/ino/parser"
)

type envPair struct {
	tyEnv  *tyEnv
	valEnv *valEnv
	next   *envPair
}

type envPairQueue struct {
	head *envPair
	last *envPair
}

func (q *envPairQueue) empty() bool {
	return q.head == nil
}

func (q *envPairQueue) enqueue(te *tyEnv, ve *valEnv) {
	p := &envPair{
		tyEnv:  te,
		valEnv: ve,
	}
	if q.head == nil {
		q.head = p
		q.last = p
		return
	}
	q.last.next = p
	q.last = p
}

func (q *envPairQueue) dequeue() *envPair {
	p := q.head
	q.head = p.next
	p.next = nil
	return p
}

type inspector struct {
	tyEnv  *tyEnv
	valEnv *valEnv
	q      *envPairQueue
}

type inspectorFunc func(symbol, *valEnvEntry, *envPair)

func (i *inspector) inspectInBreadthFirstOrder(f inspectorFunc) {
	i.q = &envPairQueue{}
	i.q.enqueue(
		&tyEnv{i.tyEnv.child},
		&valEnv{i.valEnv.child},
	)
	for !i.q.empty() {
		p := i.q.dequeue()
		if p.valEnv != nil {
			for sym, ve := range p.valEnv.bindings.m {
				if ve.tyEnv != nil || ve.valEnv != nil {
					i.q.enqueue(ve.tyEnv, ve.valEnv)
				}
				f(sym, ve, p)
			}
		}
	}
}

type typeChecker struct {
	ast    *parser.Node
	tyEnv  *tyEnv
	valEnv *valEnv
}

func (c *typeChecker) run() error {
	if err := c.resolve(); err != nil {
		return err
	}
	if err := c.check(); err != nil {
		return err
	}
	return nil
}

func (c *typeChecker) resolve() error {
	i := &inspector{
		tyEnv:  c.tyEnv,
		valEnv: c.valEnv,
	}
	i.inspectInBreadthFirstOrder(func(sym symbol, ve *valEnvEntry, p *envPair) {
		err := resolve(sym, ve, p)
		if err != nil {
			// TODO
		}
	})
	return nil
}

func (c *typeChecker) check() error {
	return checkRoot(c.ast, c.valEnv)
}

func resolve(sym symbol, ve *valEnvEntry, p *envPair) error {
	if !ve.ty.unresolved() {
		return nil
	}
	rTy, err := ve.ty.resolve(p.tyEnv)
	if err != nil {
		return fmt.Errorf("failed to resolve a type: %v: %w", sym, err)
	}
	newEntry := *ve
	newEntry.ty = rTy
	err = p.valEnv.rebind(sym, &newEntry)
	if err != nil {
		panic(fmt.Errorf("failed to re-bind symbol while resolving a type: %v: %w", sym, err))
	}
	return nil
}

func checkRoot(node *parser.Node, vEnv *valEnv) error {
	return checkDecls(node.Children[0], &valEnv{vEnv.child})
}

func checkDecls(node *parser.Node, vEnv *valEnv) error {
	for _, d := range node.Children {
		err := checkDecl(d, vEnv)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkDecl(node *parser.Node, vEnv *valEnv) error {
	switch node.Children[0].KindName {
	case "const":
		return checkConst(node.Children[0], vEnv)
	case "func":
		return checkFuncDecl(node.Children[0], vEnv)
	case "data":
		return checkDataDecl(node.Children[0], vEnv)
	}
	return fmt.Errorf("invalid declaration kind: %v", node.KindName)
}

func checkConst(node *parser.Node, vEnv *valEnv) error {
	return nil
}

func checkFuncDecl(node *parser.Node, vEnv *valEnv) error {
	funcName := node.Children[0].Text
	ee, ok := vEnv.lookup(symbol(funcName))
	if !ok {
		return fmt.Errorf("function not defined in the value environment: %v", funcName)
	}
	funcTy, ok := ee.ty.(*funcType)
	if !ok {
		return fmt.Errorf("symbol is not a function type: %v", funcName)
	}

	bodyTy, err := checkExpr(node.Children[3], ee.valEnv)
	if err != nil {
		return err
	}

	if !funcTy.result.equals(bodyTy) {
		return fmt.Errorf("body type does not match the declared type. declared type: %v, body type: %v", funcTy.result, bodyTy)
	}

	return nil
}

func checkDataDecl(node *parser.Node, vEnv *valEnv) error {
	// TODO
	return nil
}

func checkExpr(node *parser.Node, vEnv *valEnv) (declType, error) {
	switch node.Children[0].KindName {
	case "call":
		return checkCall(node.Children[0], vEnv)
	case "add":
		return checkBinaryExpr(node, vEnv)
	case "sub":
		return checkBinaryExpr(node, vEnv)
	case "mul":
		return checkBinaryExpr(node, vEnv)
	case "div":
		return checkBinaryExpr(node, vEnv)
	case "mod":
		return checkBinaryExpr(node, vEnv)
	case "id":
		return checkID(node.Children[0], vEnv)
	case "int":
		return tyInt, nil
	case "string":
		return tyString, nil
	}
	return nil, fmt.Errorf("invalid expression kind: %v", node.KindName)
}

func checkCall(node *parser.Node, vEnv *valEnv) (declType, error) {
	funcName := node.Children[0].Text
	ee, ok := vEnv.lookup(symbol(funcName))
	if !ok {
		return nil, fmt.Errorf("function is not defined: %v", funcName)
	}
	funcTy, ok := ee.ty.(*funcType)
	if !ok {
		return nil, fmt.Errorf("symbol is not a function type: %v", funcName)
	}
	args := node.Children[1].Children
	for i, arg := range args {
		argTy, err := checkExpr(arg.Children[0], vEnv)
		if err != nil {
			return nil, err
		}
		if !funcTy.params[i].equals(argTy) {
			return nil, fmt.Errorf("argument type does not match the declared type: %v's #%v argument", funcName, i)
		}
	}
	return funcTy.result, nil
}

func checkBinaryExpr(node *parser.Node, vEnv *valEnv) (declType, error) {
	funcName := "2_" + node.Children[0].Text
	ee, ok := vEnv.lookup(symbol(funcName))
	if !ok {
		return nil, fmt.Errorf("operator is not defined: %v", funcName)
	}
	funcTy, ok := ee.ty.(*funcType)
	if !ok {
		return nil, fmt.Errorf("operator is not a function: %v", funcName)
	}
	lhsTy, err := checkExpr(node.Children[1], vEnv)
	if err != nil {
		return nil, err
	}
	rhsTy, err := checkExpr(node.Children[2], vEnv)
	if err != nil {
		return nil, err
	}
	if len(funcTy.params) != 2 {
		return nil, fmt.Errorf("operator takes %v parameters, not a binary operator: %v", len(funcTy.params), node.Children[1].Text)
	}
	if !funcTy.params[0].equals(lhsTy) || !funcTy.params[1].equals(rhsTy) {
		return nil, fmt.Errorf("operator and operand don't agree. oeprator domain: %v, operand (LHS): %v, operand (RHS): %v", funcTy, lhsTy, rhsTy)
	}
	return funcTy.result, nil
}

func checkID(node *parser.Node, vEnv *valEnv) (declType, error) {
	id := node.Text
	ee, ok := vEnv.lookup(symbol(id))
	if !ok {
		return nil, fmt.Errorf("symbol is not defined: %v", id)
	}
	return ee.ty, nil
}
