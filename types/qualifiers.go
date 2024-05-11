package types

import "fmt"

type QualifierKind byte

const (
	NoQualifier QualifierKind = iota
	Const
	Input
	Simple
	Series
)

func (q QualifierKind) String() string {
	switch q {
	case NoQualifier:
		return ""
	case Const:
		return "const"
	case Input:
		return "input"
	case Simple:
		return "simple"
	case Series:
		return "series"
	}
	return "N/A"
}

type TypeWithQualifier struct {
	Qualifier QualifierKind
	Type      Type
}

func (twq TypeWithQualifier) Kind() TypeKind {
	return twq.Type.Kind()
}

func (twq TypeWithQualifier) String() string {
	return fmt.Sprintf("%s %s", twq.Qualifier.String(), twq.Type.String())
}

func (bt BaseType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (f functionType) QualifierKind() QualifierKind {
	return f.out.QualifierKind()
}

func (twq TypeWithQualifier) QualifierKind() QualifierKind {
	return twq.Qualifier
}

func (twq TypeWithQualifier) Unit() Type {
	return twq.Type.Unit()
}

func (twq TypeWithQualifier) Key() Type {
	return twq.Type.Key()
}

func (twq TypeWithQualifier) Value() Type {
	return twq.Type.Value()
}

func (twq TypeWithQualifier) Count() int {
	return twq.Type.Count()
}

func (twq TypeWithQualifier) Fields() []TypeWithName {
	return twq.Type.Fields()
}

func (twq TypeWithQualifier) Field(i int) TypeWithName {
	return twq.Type.Field(i)
}

func (twq TypeWithQualifier) FieldByName(name string) *TypeWithName {
	return twq.Type.FieldByName(name)
}

func (twq TypeWithQualifier) Items() []Type {
	return twq.Type.Items()
}

func (twq TypeWithQualifier) Item(i int) Type {
	return twq.Type.Item(i)
}

func (twq TypeWithQualifier) Members() []Type {
	return twq.Type.Members()
}

func Peel(t Type) Type {
	twq, ok := t.(TypeWithQualifier)
	if !ok {
		return t
	}

	return Peel(twq.Type)
}
