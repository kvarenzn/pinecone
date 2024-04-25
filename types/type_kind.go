package types

type TypeKind byte

const (
	UncertainKind TypeKind = iota
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
