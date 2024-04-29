package types

type TypeKind byte

const (
	NotApplicableKind TypeKind = iota
	UncertainKind
	VoidKind
	BoolKind
	IntKind
	FloatKind
	BoxKind
	ColorKind
	PointKind
	LabelKind
	LineKind
	LineFillKind
	PolyLineKind
	TableKind
	maxPrimitiveType

	// type with parameters
	ArrayKind
	MatrixKind
	MapKind
	StructKind // User Defined Types
	TupleKind // Used as function return value
	FunctionKind

	maxTypeKind
)

func (tk TypeKind) IsValid() bool {
	return tk > UncertainKind && tk < maxTypeKind
}

func (tk TypeKind) IsNormal() bool {
	return tk > VoidKind && tk < FunctionKind
}

func (tk TypeKind) IsPrimitiveType() bool {
	return tk > UncertainKind && tk < maxPrimitiveType
}
