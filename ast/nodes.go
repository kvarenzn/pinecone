package ast

import (
	"github.com/kvarenzn/pinecone/metainfo"
	"github.com/kvarenzn/pinecone/types"
)

type Node interface {
	Range() (metainfo.Location, metainfo.Location)
	SetRange(metainfo.Location, metainfo.Location)
	Begin() metainfo.Location
	SetBegin(metainfo.Location)
	End() metainfo.Location
	SetEnd(metainfo.Location)
	NodeType() types.Type
	MarkNodeType(types.Type)
	Parent() Node
	SetParent(Node)
	PathAttribute() string
	PathIndex() int
	SetPathAttribute(string)
	SetPathIndex(int)
}

type node struct {
	begin       metainfo.Location
	end         metainfo.Location
	nodeType    types.Type
	parent      Node
	rpAttribute string
	rpIndex     int
}

func (n *node) Range() (metainfo.Location, metainfo.Location) {
	return n.begin, n.end
}

func WithRange(node Node, begin, end metainfo.Location) Node {
	node.SetRange(begin, end)
	return node
}

func (n *node) SetRange(begin, end metainfo.Location) {
	n.begin = begin
	n.end = end
}

func (n *node) Begin() metainfo.Location {
	return n.begin
}

func (n *node) SetBegin(loc metainfo.Location) {
	n.begin = loc
}

func (n *node) End() metainfo.Location {
	return n.end
}

func (n *node) SetEnd(loc metainfo.Location) {
	n.end = loc
}

func (n *node) NodeType() types.Type {
	return n.nodeType
}

func (n *node) MarkNodeType(t types.Type) {
	n.nodeType = t
}

func (n *node) Parent() Node {
	return n.parent
}

func (n *node) SetParent(parent Node) {
	n.parent = parent
}

func (n *node) PathAttribute() string {
	return n.rpAttribute
}

func (n *node) PathIndex() int {
	return n.rpIndex
}

func (n *node) SetPathAttribute(name string) {
	n.rpAttribute = name
}

func (n *node) SetPathIndex(i int) {
	n.rpIndex = i
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

type BoolLiteral struct {
	node
	Value bool
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
	Default   Node
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
	Default Node
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

type Quote struct {
	node
	Content Node
}
