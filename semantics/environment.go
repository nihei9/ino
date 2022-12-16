package semantics

import (
	"fmt"
	"io"
	"strings"

	"github.com/nihei9/ino/parser"
)

type symbol string

type declType interface {
	fmt.Stringer
	unresolved() bool
	resolve(*tyEnv) (declType, error)
	equals(declType) bool
}

var (
	_ declType = basicType("")
	_ declType = &funcType{}
	_ declType = &dataType{}
	_ declType = &unresolvedType{}
)

type basicType string

func (t basicType) String() string {
	return string(t)
}

func (t basicType) unresolved() bool {
	return false
}

func (t basicType) resolve(_ *tyEnv) (declType, error) {
	return t, nil
}

func (t basicType) equals(u declType) bool {
	v, ok := u.(basicType)
	if !ok {
		return false
	}
	return t == v
}

const (
	tyInt    basicType = "int"
	tyString basicType = "string"
)

type funcType struct {
	params []declType
	result declType
}

func (t *funcType) String() string {
	var b strings.Builder
	for _, p := range t.params {
		fmt.Fprintf(&b, "%v ", p)
	}
	fmt.Fprintf(&b, "-> %v", t.result)
	return b.String()
}

func (t *funcType) unresolved() bool {
	for _, pTy := range t.params {
		if _, unresolved := pTy.(*unresolvedType); unresolved {
			return true
		}
	}
	if _, unresolved := t.result.(*unresolvedType); unresolved {
		return true
	}
	return false
}

func (t *funcType) resolve(tyEnv *tyEnv) (declType, error) {
	fTy := &funcType{
		params: make([]declType, len(t.params)),
	}
	for i, pTy := range t.params {
		if pTy.unresolved() {
			pt, err := pTy.resolve(tyEnv)
			if err != nil {
				return nil, err
			}
			fTy.params[i] = pt
		} else {
			fTy.params[i] = pTy
		}
	}
	if t.result.unresolved() {
		rt, err := t.result.resolve(tyEnv)
		if err != nil {
			return nil, err
		}
		fTy.result = rt
	} else {
		fTy.result = t.result
	}
	return fTy, nil
}

func (t *funcType) equals(u declType) bool {
	v, ok := u.(*funcType)
	if !ok {
		return false
	}
	if len(t.params) != len(v.params) {
		return false
	}
	for i, p := range t.params {
		if !p.equals(v.params[i]) {
			return false
		}
	}
	if !t.result.equals(v.result) {
		return false
	}
	return true
}

type dataType struct {
	name symbol
}

func (t *dataType) String() string {
	return string(t.name)
}

func (t *dataType) unresolved() bool {
	return false
}

func (t *dataType) resolve(_ *tyEnv) (declType, error) {
	return t, nil
}

func (t *dataType) equals(u declType) bool {
	v, ok := u.(*dataType)
	if !ok {
		return false
	}
	// TODO: タグの名前と型も比較する。
	return t.name == v.name
}

// unresolvedType is a type whose use precedes a definition in the source code.
type unresolvedType struct {
	name symbol
}

func (t *unresolvedType) String() string {
	return "?<" + string(t.name) + ">"
}

func (t *unresolvedType) unresolved() bool {
	return true
}

func (t *unresolvedType) resolve(tyEnv *tyEnv) (declType, error) {
	ty, ok := tyEnv.lookup(t.name)
	if !ok {
		return nil, fmt.Errorf("type not defined: %v", t.name)
	}
	return ty, nil
}

func (t *unresolvedType) equals(u declType) bool {
	panic("unresolvedType's equals method is called")
}

type bindings[T any] struct {
	m map[symbol]T
}

func (b *bindings[T]) bind(sym symbol, e T) bool {
	if b.m == nil {
		b.m = map[symbol]T{}
	} else {
		if _, bound := b.m[sym]; bound {
			return false
		}
	}
	b.m[sym] = e
	return true
}

func (b *bindings[T]) rebind(sym symbol, e T) error {
	if _, bound := b.m[sym]; !bound {
		return fmt.Errorf("symbol is not bound yet: %v", sym)
	}
	b.m[sym] = e
	return nil
}

func (b *bindings[T]) lookup(sym symbol) (T, bool) {
	e, ok := b.m[sym]
	return e, ok
}

type bindingsTree[T any] struct {
	*bindings[T]
	parent  *bindingsTree[T]
	child   *bindingsTree[T]
	brother *bindingsTree[T]
}

func (t *bindingsTree[T]) lookup(sym symbol) (entry T, found bool) {
	if e, ok := t.bindings.lookup(sym); ok {
		return e, true
	}
	if t.parent == nil {
		return
	}
	return t.parent.lookup(sym)
}

func (t *bindingsTree[T]) branch() *bindingsTree[T] {
	b := &bindingsTree[T]{
		bindings: &bindings[T]{},
		parent:   t,
	}
	if t.child == nil {
		t.child = b
	} else {
		b.brother = t.child
		t.child = b
	}
	return b
}

func (t *bindingsTree[T]) walk(f func(*bindingsTree[T])) {
	f(t)
	for c := t.child; c != nil; c = c.brother {
		c.walk(f)
	}
	return
}

func (t *bindingsTree[T]) inspect(f func(symbol, T, *bindingsTree[T])) {
	for sym, b := range t.bindings.m {
		f(sym, b, t)
	}
}

func (t *bindingsTree[T]) print(w io.Writer) {
	prefix := ""
	t.walk(func(tree *bindingsTree[T]) {
		tree.inspect(func(sym symbol, elem T, _ *bindingsTree[T]) {
			fmt.Printf("%v%v → %v\n", prefix, sym, elem)
		})
		prefix = prefix + "...."
	})
}

type tyEnv struct {
	*bindingsTree[declType]
}

func (e *tyEnv) lookupTentatively(sym symbol) declType {
	ty, ok := e.lookup(sym)
	if ok {
		return ty
	}
	return &unresolvedType{
		name: sym,
	}
}

type valEnvEntry struct {
	ty       declType
	constant bool
	tyEnv    *tyEnv
	valEnv   *valEnv
}

func (e *valEnvEntry) String() string {
	if e.constant {
		return "[const] " + e.ty.String()
	}
	return e.ty.String()
}

type valEnv struct {
	*bindingsTree[*valEnvEntry]
}

func (e *valEnv) lookupEnv(sym symbol) (*tyEnv, *valEnv, bool) {
	ee, ok := e.bindings.lookup(sym)
	if !ok {
		return nil, nil, false
	}
	return ee.tyEnv, ee.valEnv, true
}

type continuableError struct {
	e error
}

var _ error = &continuableError{}

func newContinuableError(e error) *continuableError {
	return &continuableError{e: e}
}

func (e *continuableError) Error() string {
	return e.e.Error()
}

type environmentBuilder struct {
	tyEnvRoot  *tyEnv
	valEnvRoot *valEnv
	tyEnv      *tyEnv
	valEnv     *valEnv
	errs       []error
}

func (b *environmentBuilder) run(root *parser.Node) (retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = err.(error)
		}
	}()

	tyEnv := &tyEnv{
		bindingsTree: &bindingsTree[declType]{
			bindings: &bindings[declType]{},
		},
	}
	valEnv := &valEnv{
		bindingsTree: &bindingsTree[*valEnvEntry]{
			bindings: &bindings[*valEnvEntry]{},
		},
	}
	b.tyEnvRoot = tyEnv
	b.valEnvRoot = valEnv
	b.tyEnv = tyEnv
	b.valEnv = valEnv
	b.buildIn()
	b.buildRoot(root)
	if len(b.errs) > 0 {
		return fmt.Errorf("failed to build an envrironment")
	}
	return nil
}

func (b *environmentBuilder) buildIn() {
	b.tyEnv.bind("int", tyInt)
	b.tyEnv.bind("string", tyString)

	b.valEnv.bind("2_+", &valEnvEntry{
		ty: &funcType{
			params: []declType{tyInt, tyInt},
			result: tyInt,
		},
	})
	b.valEnv.bind("2_-", &valEnvEntry{
		ty: &funcType{
			params: []declType{tyInt, tyInt},
			result: tyInt,
		},
	})
	b.valEnv.bind("2_*", &valEnvEntry{
		ty: &funcType{
			params: []declType{tyInt, tyInt},
			result: tyInt,
		},
	})
	b.valEnv.bind("2_/", &valEnvEntry{
		ty: &funcType{
			params: []declType{tyInt, tyInt},
			result: tyInt,
		},
	})
	b.valEnv.bind("2_%", &valEnvEntry{
		ty: &funcType{
			params: []declType{tyInt, tyInt},
			result: tyInt,
		},
	})
}

func (b *environmentBuilder) enter() (*tyEnv, *valEnv) {
	t := b.tyEnv.bindingsTree
	v := b.valEnv.bindingsTree
	b.tyEnv.bindingsTree = b.tyEnv.branch()
	b.valEnv.bindingsTree = b.valEnv.branch()
	return &tyEnv{t}, &valEnv{v}
}

func (b *environmentBuilder) leave() {
	if b.tyEnv.bindingsTree.parent == nil || b.valEnv.bindingsTree.parent == nil {
		b.fatal("invalid operation: leave()")
	}
	b.tyEnv.bindingsTree = b.tyEnv.bindingsTree.parent
	b.valEnv.bindingsTree = b.valEnv.bindingsTree.parent
}

func (b *environmentBuilder) error(format string, a ...any) {
	err := newContinuableError(fmt.Errorf(format, a...))
	b.errs = append(b.errs, err)
	panic(err)
}

func (b *environmentBuilder) fatal(format string, a ...any) {
	panic(fmt.Errorf(format, a...))
}

func (b *environmentBuilder) buildRoot(node *parser.Node) {
	b.enter()
	defer b.leave()
	b.buildDecls(node.Children[0])
}

func (b *environmentBuilder) buildDecls(node *parser.Node) {
	for _, c := range node.Children {
		b.buildDecl(c)
	}
}

func (b *environmentBuilder) buildDecl(node *parser.Node) {
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(*continuableError); ok {
				return
			}
			panic(err)
		}
	}()

	d := node.Children[0]
	switch d.KindName {
	case "const":
		b.buildConst(d)
	case "func":
		b.buildFunc(d)
	case "data":
		b.buildData(d)
	default:
		b.fatal("invalid node kind: %v", d.KindName)
	}
}

func (b *environmentBuilder) buildConst(node *parser.Node) {
	name := node.Children[0].Text

	var ty declType
	{
		tyLit := node.Children[1].Children[0]
		tySym := symbol(tyLit.Children[0].Text)
		ty = b.tyEnv.lookupTentatively(tySym)
	}

	_, ve := b.enter()
	defer b.leave()
	ok := ve.bind(symbol(name), &valEnvEntry{
		constant: true,
		ty:       ty,
	})
	if !ok {
		b.error("duplicated symbol: %v", name)
	}

	b.buildExpr(node.Children[2])
}

func (b *environmentBuilder) buildFunc(node *parser.Node) {
	var paramSyms []symbol
	var paramTys []declType
	{
		params := node.Children[1]
		paramSyms = make([]symbol, len(params.Children))
		paramTys = make([]declType, len(params.Children))
		for i, p := range params.Children {
			paramSyms[i] = symbol(p.Children[0].Text)
			pTyLit := p.Children[1].Children[0]
			tySym := symbol(pTyLit.Children[0].Text)
			paramTys[i] = b.tyEnv.lookupTentatively(tySym)
		}
	}
	var resultTy declType
	{
		rTyLit := node.Children[2].Children[0]
		tySym := symbol(rTyLit.Children[0].Text)
		resultTy = b.tyEnv.lookupTentatively(tySym)
	}

	_, ve := b.enter()
	defer b.leave()
	funcName := node.Children[0].Text
	ok := ve.bind(symbol(funcName), &valEnvEntry{
		ty: &funcType{
			params: paramTys,
			result: resultTy,
		},
		tyEnv:  &tyEnv{b.tyEnv.bindingsTree},
		valEnv: &valEnv{b.valEnv.bindingsTree},
	})
	if !ok {
		b.error("duplicated symbol: %v", funcName)
	}

	for i, sym := range paramSyms {
		ok := b.valEnv.bind(sym, &valEnvEntry{
			ty: paramTys[i],
		})
		if !ok {
			b.error("duplicated symbol: %v", sym)
		}
	}

	b.buildExpr(node.Children[3])
}

func (b *environmentBuilder) buildData(node *parser.Node) {
	resultTy := &dataType{
		name: symbol(node.Children[0].Text),
	}
	ok := b.tyEnv.bind(resultTy.name, resultTy)
	if !ok {
		b.error("duplicated symbol: %v", resultTy.name)
	}
	for _, cons := range node.Children[1].Children {
		var paramTys []declType
		{
			optTyLits := cons.Children[1]
			if len(optTyLits.Children) > 0 {
				tyLits := optTyLits.Children[0]
				paramTys = make([]declType, len(tyLits.Children))
				for i, tyLit := range tyLits.Children {
					tySym := symbol(tyLit.Children[0].Text)
					paramTys[i] = b.tyEnv.lookupTentatively(tySym)
				}
			}
		}
		tagName := cons.Children[0].Text
		ok := b.valEnv.bind(symbol(tagName), &valEnvEntry{
			ty: &funcType{
				params: paramTys,
				result: resultTy,
			},
		})
		if !ok {
			b.error("duplicated symbol: %v", tagName)
		}
	}
}

func (b *environmentBuilder) buildExpr(node *parser.Node) {
	// TODO
}
