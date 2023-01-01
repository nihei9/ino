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
	Name     string
	TypeVars []int
}

type TypeVar struct {
	Num int
}

type Decl interface {
}

var (
	_ Decl = &DataDecl{}
	_ Decl = &ValConsDecl{}
)

type DataDecl struct {
	Name         string
	TypeVarCount int
}

type ValConsDecl struct {
	Name         string
	TyName       string
	Params       []Type
	TypeVarCount int
}

type File struct {
	Decls []Decl
}
