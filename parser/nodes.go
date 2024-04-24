package parser

import "github.com/kvarenzn/pinecone/tokenizer"

type Type interface {
	_type()
}

type SimpleType struct {
	Name tokenizer.Token
}

type SubType struct {
	Name   Type
	Member tokenizer.Token
}

type GenericType struct {
	Name Type
	Args []Type
}

func (st SimpleType) _type()  {}
func (st SubType) _type()     {}
func (gt GenericType) _type() {}

type Expr interface {
	_expr()
}

type BinaryExpr struct {
	Left  Expr
	Op    tokenizer.Token
	Right Expr
}

type UnaryExpr struct {
	Op   tokenizer.Token
	Expr Expr
}

type AttrExpr struct {
	Target Expr
	Name   tokenizer.Token
}

type KwArg struct {
	Name  tokenizer.Token
	Value Expr
}

type InstantiationExpr struct {
	Template Expr
	TypeArgs []Type
}

type CallExpr struct {
	Func Expr
	Args []Expr
}

type HRefExpr struct {
	Series Expr
	Offset Expr
}

type Identifier struct {
	Name tokenizer.Token
}

type StringLiteral struct {
	Value string
}

type IntLiteral struct {
	Value int64
}

type FloatLiteral struct {
	Value float64
}

type ColorLiteral struct {
	R float64
	G float64
	B float64
	T float64
}

type TrueExpr struct{}

type FalseExpr struct{}

type TupleExpr struct {
	Items []Expr
}

type TernaryExpr struct {
	Test  Expr
	True  Expr
	False Expr
}

func (be BinaryExpr) _expr()        {}
func (ue UnaryExpr) _expr()         {}
func (ae AttrExpr) _expr()          {}
func (ka KwArg) _expr()             {}
func (ie InstantiationExpr) _expr() {}
func (ce CallExpr) _expr()          {}
func (he HRefExpr) _expr()          {}
func (i Identifier) _expr()         {}
func (sl StringLiteral) _expr()     {}
func (il IntLiteral) _expr()        {}
func (fl FloatLiteral) _expr()      {}
func (cl ColorLiteral) _expr()      {}
func (te TrueExpr) _expr()          {}
func (fe FalseExpr) _expr()         {}
func (te TupleExpr) _expr()         {}
func (te TernaryExpr) _expr()       {}

type Stmt interface {
	_stmt()
}

type ExprStmt struct {
	Expr Expr
}

type VarDeclStmt struct {
	DeclMode  *tokenizer.Token
	Qualifier *tokenizer.Token
	Type      Type
	Name      tokenizer.Token
	Initial   Stmt
}

type TupleDeclStmt struct {
	Variables []tokenizer.Token
	Initial   Stmt
}

type ReassignStmt struct {
	Target Expr
	Op     tokenizer.Token
	Value  Stmt
}

type IfStmt struct {
	Test  Expr
	True  Stmt
	False Stmt
}

type CaseClause struct {
	Cond Expr
	Body Stmt
}

type SwitchStmt struct {
	Target  Expr
	Cases   []CaseClause
	Default Stmt
}

type WhileStmt struct {
	Test Expr
	Body Stmt
}

type ForStmt struct {
	Counter tokenizer.Token
	Init    Expr
	Step    Expr
	Final   Expr
	Body    Stmt
}

type ForInStmt struct {
	Index     tokenizer.Token
	Iterator  tokenizer.Token
	Container Expr
	Body      Stmt
}

type BreakStmt struct{}
type ContinueStmt struct{}

type ParamDecl struct {
	Qualifier *tokenizer.Token
	Type      Type
	Name      tokenizer.Token
	Default   *tokenizer.Token
}

type FuncDeclStmt struct {
	Export bool
	Method bool
	Name   tokenizer.Token
	Params []ParamDecl
	Body   Stmt
}

type MemberDecl struct {
	Type    Type
	Name    tokenizer.Token
	Default *tokenizer.Token
}

type TypeDeclStmt struct {
	Name    tokenizer.Token
	Members []MemberDecl
}

type ImportStmt struct {
	User    tokenizer.Token
	Name    tokenizer.Token
	Version tokenizer.Token
	Alias   *tokenizer.Token
}

type Suite struct {
	Body []Stmt
}

func (es ExprStmt) _stmt()       {}
func (vds VarDeclStmt) _stmt()   {}
func (tds TupleDeclStmt) _stmt() {}
func (rs ReassignStmt) _stmt()   {}
func (is IfStmt) _stmt()         {}
func (cc CaseClause) _stmt()     {}
func (ss SwitchStmt) _stmt()     {}
func (ws WhileStmt) _stmt()      {}
func (fs ForStmt) _stmt()        {}
func (fis ForInStmt) _stmt()     {}
func (bs BreakStmt) _stmt()      {}
func (cs ContinueStmt) _stmt()   {}
func (pd ParamDecl) _stmt()      {}
func (fds FuncDeclStmt) _stmt()  {}
func (md MemberDecl) _stmt()     {}
func (tds TypeDeclStmt) _stmt()  {}
func (is ImportStmt) _stmt()     {}
func (s Suite) _stmt()           {}
