package ir

type Type interface {
}

var (
	_ Type = &BasicType{}
	_ Type = &NamedType{}
	_ Type = &TypeVar{}
)

type BasicType struct {
	Name string
}

type NamedType struct {
	Name string
}

type ConcreteType struct {
	AbstractTy Type
	Args       []Type
}

type TypeVar struct {
	Num int
}

type Decl interface {
}

var (
	_ Decl = &DataDecl{}
)

type DataDecl struct {
	Name         string
	TypeVarCount int
	Conss        []*ValCons
}

type ValCons struct {
	Name   string
	Params []Type
}

type File struct {
	Decls []Decl
}
