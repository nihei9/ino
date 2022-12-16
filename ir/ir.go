package ir

type Type interface {
}

var (
	_ Type = &BasicType{}
	_ Type = &NamedType{}
)

type BasicType struct {
	Name string
}

type NamedType struct {
	Name string
}

type Decl interface {
}

var (
	_ Decl = &FuncDecl{}
	_ Decl = &DataDecl{}
	_ Decl = &ValConsDecl{}
)

type BinOp string

const (
	OpAdd BinOp = "+"
	OpSub BinOp = "-"
	OpMul BinOp = "*"
	OpDiv BinOp = "/"
	OpMod BinOp = "%"
)

type Expr interface {
}

var (
	_ Expr = &IntLit{}
	_ Expr = &StringLit{}
	_ Expr = &IdentExpr{}
	_ Expr = &CallExpr{}
	_ Expr = &BinaryExpr{}
)

type IntLit struct {
	Val int
}

type StringLit struct {
	Val string
}

type IdentExpr struct {
	Ident string
	Constant  bool
}

type CallExpr struct {
	Func Expr
	Args []Expr
}

type BinaryExpr struct {
	LHS Expr
	Op  BinOp
	RHS Expr
}

type Param struct {
	Name string
	Ty   Type
}

type FuncDecl struct {
	Name   string
	Params []*Param
	Result Type
	Body   Expr
}

type DataDecl struct {
	Name string
}

type ValConsDecl struct {
	Name   string
	TyName string
	Params []Type
}

type File struct {
	Decls []Decl
}
