package types

type Type interface {
	Kind() TypeKind
	String() string
	QualifierKind() QualifierKind
}
