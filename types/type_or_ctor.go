package types

import "fmt"

type TocKind byte

const (
	TocType TocKind = iota
	TocCtor
)

type TypeOrCtor struct {
	BaseType
	Tag  TocKind
	Type Type
	Ctor func(args []Type) (Type, error)
}

func NewTocType(t Type) TypeOrCtor {
	return TypeOrCtor{
		Tag: TocType,
		Type: t,
	}
}

func NewTocCtor(ctor func(args []Type) (Type, error)) TypeOrCtor {
	return TypeOrCtor{
		Tag: TocCtor,
		Ctor: ctor,
	}
}

func (toc TypeOrCtor) Kind() TypeKind {
	return TypeOrCtorKind
}

func (toc TypeOrCtor) String() string {
	return fmt.Sprintf("toc(%v)", toc.Tag)
}
