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


// QualifierKind()
func (u uncertainType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (v voidType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (b boolType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (i intType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (f floatType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (b boxType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (c colorType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (p pointType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (l labelType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (l lineType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (l lineFillType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (p polyLineType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (t tableType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (m mapType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (a arrayType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (m matrixType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (s structType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (t tupleType) QualifierKind() QualifierKind {
	return NoQualifier
}

func (f functionType) QualifierKind() QualifierKind {
	return f.out.QualifierKind()
}

func (twq TypeWithQualifier) QualifierKind() QualifierKind {
	return twq.Qualifier
}

