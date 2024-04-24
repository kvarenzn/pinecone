package ast

import "github.com/kvarenzn/pinecone/structs"

type Node interface {
	Range() (structs.Location, structs.Location)
	SetRange(structs.Location, structs.Location)
	Begin() structs.Location
	SetBegin(structs.Location)
	End() structs.Location
	SetEnd(structs.Location)
}

type node struct {
	begin structs.Location
	end   structs.Location
}

func (n *node) Range() (structs.Location, structs.Location) {
	return n.begin, n.end
}

func WithRange(node Node, begin, end structs.Location) Node {
	node.SetRange(begin, end)
	return node
}

func (n *node) SetRange(begin, end structs.Location) {
	n.begin = begin
	n.end = end
}

func (n *node) Begin() structs.Location {
	return n.begin
}

func (n *node) SetBegin(loc structs.Location) {
	n.begin = loc
}

func (n *node) End() structs.Location {
	return n.end
}

func (n *node) SetEnd(loc structs.Location) {
	n.end = loc
}

type SimpleType struct {
	node
	Name string
}

type SubType struct {
	node
	Name   Node
	Member string
}

type GenericType struct {
	node
	Name Node
	Args []Node
}

type BinaryExpr struct {
	node
	Left  Node
	Op    string
	Right Node
}

type UnaryExpr struct {
	node
	Op   string
	Expr Node
}

type AttrExpr struct {
	node
	Target Node
	Name   string
}

type KwArg struct {
	node
	Name  string
	Value Node
}

type InstantiationExpr struct {
	node
	Template Node
	TypeArgs []Node
}

type CallExpr struct {
	node
	Func Node
	Args []Node
}

type HRefExpr struct {
	node
	Series Node
	Offset Node
}

type Identifier struct {
	node
	Name string
}

type StringLiteral struct {
	node
	Value string
}

type IntLiteral struct {
	node
	Value int64
}

type FloatLiteral struct {
	node
	Value float64
}

type ColorLiteral struct {
	node
	R float64
	G float64
	B float64
	T float64
}

type TrueExpr struct {
	node
}

type FalseExpr struct {
	node
}

type TupleExpr struct {
	node
	Items []Node
}

type TernaryExpr struct {
	node
	Test  Node
	True  Node
	False Node
}

type ExprStmt struct {
	node
	Expr Node
}

type VarDeclStmt struct {
	node
	DeclMode  *string
	Qualifier *string
	Type      Node
	Name      string
	Initial   Node
}

type TupleDeclStmt struct {
	node
	Variables []string
	Initial   Node
}

type ReassignStmt struct {
	node
	Target Node
	Op     string
	Value  Node
}

type IfStmt struct {
	node
	Test  Node
	True  Node
	False Node
}

type CaseClause struct {
	node
	Cond Node
	Body Node
}

type SwitchStmt struct {
	node
	Target  Node
	Cases   []*CaseClause
	Default Node
}

type WhileStmt struct {
	node
	Test Node
	Body Node
}

type ForStmt struct {
	node
	Counter string
	Init    Node
	Step    Node
	Final   Node
	Body    Node
}

type ForInStmt struct {
	node
	Index     *string
	Iterator  string
	Container Node
	Body      Node
}

type BreakStmt struct {
	node
}
type ContinueStmt struct {
	node
}

type ParamDecl struct {
	node
	Qualifier *string
	Type      Node
	Name      string
	Default   *string
}

type FuncDeclStmt struct {
	node
	Export bool
	Method bool
	Name   string
	Params []*ParamDecl
	Body   Node
}

type MemberDecl struct {
	node
	Type    Node
	Name    string
	Default *string
}

type TypeDeclStmt struct {
	node
	Name    string
	Members []*MemberDecl
}

type ImportStmt struct {
	node
	User    string
	Name    string
	Version string
	Alias   *string
}

type Suite struct {
	node
	Body []Node
}
