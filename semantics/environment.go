package semantics

import (
	"fmt"
	"io"
	"strings"

	"github.com/nihei9/ino/parser"
)

type symbol string

type abstractType interface {
	tyParams() []symbol
}

var _ abstractType = &dataType{}

type declType interface {
	fmt.Stringer
	abstract() (abstractType, bool)
	unresolved() bool
	resolve(*tyEnv) (declType, error)
	equals(declType) bool
}

var (
	_ declType = basicType("")
	_ declType = &funcType{}
	_ declType = &dataType{}
	_ declType = &concreteType{}
	_ declType = &typeVar{}
	_ declType = &unresolvedType{}
)

type basicType string

func (t basicType) String() string {
	return string(t)
}

func (t basicType) abstract() (abstractType, bool) {
	return nil, false
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
	tyInt    basicType = "Int"
	tyString basicType = "String"
)

type funcType struct {
	tyVars []symbol
	params []declType
	result declType
}

func (t *funcType) String() string {
	var b strings.Builder
	if len(t.params) > 0 {
		for _, p := range t.params {
			fmt.Fprintf(&b, "%v ", p)
		}
	} else {
		fmt.Fprint(&b, "() ")
	}
	fmt.Fprintf(&b, "-> %v", t.result)
	return b.String()
}

func (t *funcType) abstract() (abstractType, bool) {
	return nil, false
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
		tyVars: t.tyVars,
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
	return t.result.equals(v.result)
}

type dataType struct {
	name   symbol
	tyVars []symbol
}

func (t *dataType) String() string {
	if len(t.tyVars) == 0 {
		return string(t.name)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "(%v", t.name)
	for _, v := range t.tyVars {
		fmt.Fprintf(&b, " %v", v)
	}
	fmt.Fprintf(&b, ")")
	return b.String()
}

func (t *dataType) abstract() (abstractType, bool) {
	if len(t.tyVars) == 0 {
		return nil, false
	}
	return t, true
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
	return t.name == v.name
}

func (t *dataType) tyParams() []symbol {
	return t.tyVars
}

type concreteType struct {
	abstractTy declType
	args       []declType
}

func (t *concreteType) String() string {
	var b strings.Builder
	fmt.Fprint(&b, t.abstractTy)
	fmt.Fprintf(&b, "<%v", t.args[0])
	for _, p := range t.args[1:] {
		fmt.Fprintf(&b, " %v", p)
	}
	fmt.Fprint(&b, ">")
	return b.String()
}

func (t *concreteType) abstract() (abstractType, bool) {
	return nil, false
}

func (t *concreteType) unresolved() bool {
	if t.abstractTy.unresolved() {
		return true
	}
	for _, p := range t.args {
		if p.unresolved() {
			return true
		}
	}
	return false
}

func (t *concreteType) resolve(tyEnv *tyEnv) (declType, error) {
	cTy := &concreteType{
		abstractTy: t.abstractTy,
		args:       make([]declType, len(t.args)),
	}
	if t.abstractTy.unresolved() {
		ty, err := t.abstractTy.resolve(tyEnv)
		if err != nil {
			return nil, err
		}
		cTy.abstractTy = ty
	}
	for i, p := range t.args {
		if p.unresolved() {
			ty, err := p.resolve(tyEnv)
			if err != nil {
				return nil, err
			}
			cTy.args[i] = ty
		} else {
			cTy.args[i] = p
		}
	}
	return cTy, nil
}

func (t *concreteType) equals(u declType) bool {
	v, ok := u.(*concreteType)
	if !ok {
		return false
	}
	if !t.abstractTy.equals(v.abstractTy) {
		return false
	}
	for i, p := range t.args {
		if !p.equals(v.args[i]) {
			return false
		}
	}
	return true
}

type typeVar struct {
	name symbol
}

func (t *typeVar) String() string {
	return string(t.name)
}

func (t *typeVar) abstract() (abstractType, bool) {
	return nil, false
}

func (t *typeVar) unresolved() bool {
	return false
}

func (t *typeVar) resolve(tyEnv *tyEnv) (declType, error) {
	return t, nil
}

func (t *typeVar) equals(u declType) bool {
	v, ok := u.(*typeVar)
	if !ok {
		return false
	}
	return t.name == v.name
}

// unresolvedType is a type whose use precedes a definition in the source code.
type unresolvedType struct {
	name symbol
}

func (t *unresolvedType) String() string {
	return "?<" + string(t.name) + ">"
}

func (t *unresolvedType) abstract() (abstractType, bool) {
	return nil, false
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
			fmt.Fprintf(w, "%v%v → %v\n", prefix, sym, elem)
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
	ty     declType
	tyEnv  *tyEnv
	valEnv *valEnv
}

func (e *valEnvEntry) String() string {
	return e.ty.String()
}

type valEnv struct {
	*bindingsTree[*valEnvEntry]
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
	b.tyEnv.bind("Int", tyInt)
	b.tyEnv.bind("String", tyString)
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
	case "data":
		b.buildData(d)
	default:
		b.fatal("invalid node kind: %v", d.KindName)
	}
}

func (b *environmentBuilder) buildData(node *parser.Node) {
	tyVars := make([]symbol, len(node.Children[1].Children))
	for i, tv := range node.Children[1].Children {
		tyVars[i] = symbol(tv.Children[0].Text)
	}
	resultTy := &dataType{
		name:   symbol(node.Children[0].Text),
		tyVars: tyVars,
	}
	ok := b.tyEnv.bind(resultTy.name, resultTy)
	if !ok {
		b.error("duplicated symbol: %v", resultTy.name)
	}
	for _, cons := range node.Children[2].Children {
		var paramTys []declType
		{
			optTyLits := cons.Children[1]
			if len(optTyLits.Children) > 0 {
				tyLits := optTyLits.Children[0]
				paramTys = make([]declType, len(tyLits.Children))
				for i, tyLit := range tyLits.Children {
					var ty declType
					{
						tySym := symbol(tyLit.Children[0].Text)
						for _, v := range tyVars {
							if tySym == v {
								ty = &typeVar{
									name: tySym,
								}
							}
						}
						if ty == nil {
							ty = b.tyEnv.lookupTentatively(tySym)
						}
					}
					cTy, err := concretize(ty, tyLit, b.tyEnv)
					if err != nil {
						b.error(err.Error())
						return
					}
					paramTys[i] = cTy
				}
			}
		}
		tagName := cons.Children[0].Text
		ok := b.valEnv.bind(symbol(tagName), &valEnvEntry{
			ty: &funcType{
				tyVars: tyVars,
				params: paramTys,
				result: resultTy,
			},
		})
		if !ok {
			b.error("duplicated symbol: %v", tagName)
		}
	}
}

// FIXME: Move the code for type checking to `type_checker.go`.
func concretize(ty declType, tyLit *parser.Node, te *tyEnv) (declType, error) {
	abTy, ok := ty.abstract()
	if !ok {
		return ty, nil
	}
	tyParams := abTy.tyParams()
	// Check whether ty_lit has ty_vars
	if len(tyLit.Children) < 2 {
		return nil, fmt.Errorf("failed to concretize: want %v arguments but got no arguments", len(tyParams))
	}
	tyArgs := tyLit.Children[1]
	if len(tyParams) != len(tyArgs.Children) {
		return nil, fmt.Errorf("failed to concretize: want %v arguments but got %v arguments", len(tyParams), len(tyArgs.Children))
	}
	args := make([]declType, len(tyArgs.Children))
	for i, tyArg := range tyArgs.Children {
		argSym := symbol(tyArg.Children[0].Text)
		var t declType
		for _, paramSym := range tyParams {
			if paramSym == argSym {
				t = &typeVar{
					name: argSym,
				}
			}
		}
		if t == nil {
			t = te.lookupTentatively(argSym)
		}
		args[i] = t
	}
	return &concreteType{
		abstractTy: ty,
		args:       args,
	}, nil
}
