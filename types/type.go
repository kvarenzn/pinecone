package types

type Type interface {
	// common
	Kind() TypeKind
	String() string

	// qualifier
	QualifierKind() QualifierKind
	
	// containers
	Unit() Type // Array, Matrix

	// map
	Key() Type
	Value() Type

	Count() int // tuple, struct, function(argument count)

	// struct
	Fields() []TypeWithName
	Field(i int) TypeWithName
	FieldByName(name string) *TypeWithName

	// tuple
	Items() []Type
	Item(i int) Type

	// function
	In(i int) TypeWithName
	InName(name string) *TypeWithName
	AllIn() []TypeWithName
	Out() Type

	// union
	Members() []Type
}
