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
	_ Decl = &DataDecl{}
	_ Decl = &ValConsDecl{}
)

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
