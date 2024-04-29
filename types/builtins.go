package types

import (
	"fmt"
	"strings"
)

// / Definitions
type baseType struct{}

type uncertainType struct {
	baseType
}
type voidType struct {
	baseType
}
type boolType struct {
	baseType
}
type intType struct {
	baseType
}
type floatType struct {
	baseType
}
type boxType struct {
	baseType
}
type colorType struct {
	baseType
}
type pointType struct { // chart.point
	baseType
}
type labelType struct {
	baseType
}
type lineType struct {
	baseType
}
type lineFillType struct {
	baseType
}
type polyLineType struct {
	baseType
}
type tableType struct {
	baseType
}
type mapType struct {
	baseType
	key   Type
	value Type
}
type arrayType struct {
	baseType
	item Type
}
type matrixType struct {
	baseType
	unit Type
}
type TypeWithName struct {
	Name     string
	Type     Type
	Optional bool
}
type structType struct {
	baseType
	fields []TypeWithName
}
type tupleType struct {
	baseType
	items []Type
}
type functionType struct {
	baseType
	in  []TypeWithName
	out Type
}

// Constructors
var Uncertain = new(uncertainType)
var Void = new(voidType)
var Bool = new(boolType)
var Int = new(intType)
var Float = new(floatType)
var Box = new(boxType)
var Point = new(pointType)
var Label = new(labelType)
var Line = new(lineType)
var LineFill = new(lineFillType)
var PolyLine = new(polyLineType)
var Table = new(tableType)

func MapOf(key, value Type) Type {
	return mapType{
		key:   key,
		value: value,
	}
}

func ArrayOf(item Type) Type {
	return arrayType{
		item: item,
	}
}

func MatrixOf(unit Type) Type {
	return matrixType{
		unit: unit,
	}
}

func StructOf(fields []TypeWithName) Type {
	return structType{
		fields: fields,
	}
}

func TupleOf(items []Type) Type {
	return tupleType{
		items: items,
	}
}

func Tuple(items ...Type) Type {
	return TupleOf(items)
}

func FunctionOf(in []TypeWithName, out Type) Type {
	return functionType{
		in:  in,
		out: out,
	}
}

// Kind()
func (u uncertainType) Kind() TypeKind {
	return UncertainKind
}

func (v voidType) Kind() TypeKind {
	return VoidKind
}

func (b boolType) Kind() TypeKind {
	return BoolKind
}

func (i intType) Kind() TypeKind {
	return IntKind
}

func (f floatType) Kind() TypeKind {
	return FloatKind
}

func (b boxType) Kind() TypeKind {
	return BoxKind
}

func (c colorType) Kind() TypeKind {
	return ColorKind
}

func (p pointType) Kind() TypeKind {
	return PointKind
}

func (l labelType) Kind() TypeKind {
	return LabelKind
}

func (l lineType) Kind() TypeKind {
	return LineKind
}

func (lf lineFillType) Kind() TypeKind {
	return LineFillKind
}

func (pl polyLineType) Kind() TypeKind {
	return PolyLineKind
}

func (t tableType) Kind() TypeKind {
	return TableKind
}

func (m mapType) Kind() TypeKind {
	return MapKind
}

func (a arrayType) Kind() TypeKind {
	return ArrayKind
}

func (m matrixType) Kind() TypeKind {
	return MatrixKind
}

func (s structType) Kind() TypeKind {
	return StructKind
}

func (t tupleType) Kind() TypeKind {
	return TupleKind
}

func (s functionType) Kind() TypeKind {
	return FunctionKind
}

// String()
func (u uncertainType) String() string {
	return "uncertain"
}

func (v voidType) String() string {
	return "void"
}

func (b boolType) String() string {
	return "bool"
}

func (i intType) String() string {
	return "int"
}

func (f floatType) String() string {
	return "float"
}

func (b boxType) String() string {
	return "box"
}

func (c colorType) String() string {
	return "color"
}

func (p pointType) String() string {
	return "chart.point"
}

func (l labelType) String() string {
	return "label"
}

func (l lineType) String() string {
	return "line"
}

func (l lineFillType) String() string {
	return "linefill"
}

func (p polyLineType) String() string {
	return "polyline"
}

func (t tableType) String() string {
	return "table"
}

func (m mapType) String() string {
	return fmt.Sprintf("map<%s, %s>", m.key, m.value)
}

func (a arrayType) String() string {
	return fmt.Sprintf("array<%s>", a.item)
}

func (m matrixType) String() string {
	return fmt.Sprintf("matrix<%s>", m.unit)
}

func (twn TypeWithName) String() string {
	return fmt.Sprintf("%s %s", twn.Type, twn.Name)
}

func (s structType) String() string {
	fs := []string{}
	for _, f := range s.fields {
		fs = append(fs, f.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(fs, "; "))
}

func (t tupleType) String() string {
	is := []string{}
	for _, i := range t.items {
		is = append(is, i.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(is, ", "))
}

func (f functionType) String() string {
	ins := []string{}
	for _, i := range f.in {
		ins = append(ins, i.String())
	}
	return fmt.Sprintf("(%s) => %s", strings.Join(ins, ", "), f.out.String())
}

// Utilities
func CanDoImplicitConversion(from, to Type) bool {
	if to.Kind() == UncertainKind && from.Kind() != UncertainKind && from.Kind() != VoidKind {
		// na can be casted to almost any types
		return true
	}
	if to.Kind() == BoolKind {
		if from.Kind() == IntKind || from.Kind() == FloatKind {
			return true
		}
	} else if to.Kind() == FloatKind && from.Kind() == IntKind {
		return true
	}

	return false
}

// containers
func (bt baseType) Unit() Type {
	panic("not applicable")
}

func (a arrayType) Unit() Type {
	return a.item
}

func (m matrixType) Unit() Type {
	return m.unit
}

// map
func (bt baseType) Key() Type {
	panic("not applicable")
}

func (bt baseType) Value() Type {
	panic("not applicable")
}

func (m mapType) Key() Type {
	return m.key
}

func (m mapType) Value() Type {
	return m.value
}

// Count()
func (bt baseType) Count() int {
	panic("not applicable")
}

func (st structType) Count() int {
	return len(st.fields)
}

func (tt tupleType) Count() int {
	return len(tt.items)
}

// structs
func (bt baseType) Fields() []TypeWithName {
	panic("not applicable")
}

func (bt baseType) Field(i int) TypeWithName {
	panic("not applicable")
}

func (bt baseType) FieldByName(name string) *TypeWithName {
	panic("not applicable")
}

func (st structType) Fields() []TypeWithName {
	fields := []TypeWithName{}
	for _, f := range st.fields {
		fields = append(fields, f)
	}
	return fields
}

func (st structType) Field(i int) TypeWithName {
	return st.fields[i]
}

func (st structType) FieldByName(name string) *TypeWithName {
	for _, f := range st.fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

// tuple
func (bt baseType) Items() []Type {
	panic("not applicable")
}

func (tt tupleType) Items() []Type {
	result := []Type{}
	for _, i := range tt.items {
		result = append(result, i)
	}
	return result
}

func (bt baseType) Item(i int) Type {
	panic("not applicable")
}

func (tt tupleType) Item(i int) Type {
	return tt.items[i]
}
