package analyzer

import (
	"fmt"

	"github.com/kvarenzn/pinecone/ast"
	"github.com/kvarenzn/pinecone/base"
	"github.com/kvarenzn/pinecone/builtins"
	"github.com/kvarenzn/pinecone/types"
)

type typeAnalyzer struct {
	scopes    []map[string]types.TypeWithQualifier
	namespace base.Namespace
	userNS    base.Namespace
	errors    []error
}

func AnalyzeType(namespace base.Namespace, root ast.Node) []error {
	analyzer := typeAnalyzer{
		scopes:    []map[string]types.TypeWithQualifier{make(map[string]types.TypeWithQualifier)},
		namespace: namespace,
		userNS: base.Namespace{
			Callables: map[string]types.Callable{},
			Types:     map[string]types.TypeOrCtor{},
		},
		errors: []error{},
	}
	analyzer.markType(root)

	return analyzer.errors
}

func (ta typeAnalyzer) lookupType(name string) (types.Type, error) {
	// find in user-defined types
	t, err := ta.userNS.FindType(name)
	if err == nil {
		return t, nil
	}

	// find in builtin types
	t, err = ta.namespace.FindType(name)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (ta typeAnalyzer) lookupIdentifier(name string) (types.Type, error) {
	// 1. find variables
	last := len(ta.scopes) - 1
	for i := last; i >= 0; i-- {
		v, ok := ta.scopes[i][name]
		if ok {
			return v, nil
		}
	}

	// 2. find functions
	fn, err := ta.namespace.FindFunction(name)
	if err != nil {
		return types.CallableTypeWrap(fn), nil
	}

	// 3. find namespace
	m, err := ta.namespace.FindNamespace(name)
	if err == nil {
		return base.NSTypeWrap(*m), nil
	}

	return nil, fmt.Errorf("unknown identifier '%s'", name)
}

func (ta *typeAnalyzer) enterScope() {
	ta.scopes = append(ta.scopes, map[string]types.TypeWithQualifier{})
}

func (ta *typeAnalyzer) exitScope() {
	ta.scopes = ta.scopes[:len(ta.scopes)-1]
}

func (ta *typeAnalyzer) registerVariable(name string, twq types.TypeWithQualifier) error {
	last := ta.scopes[len(ta.scopes)-1]
	if _, ok := last[name]; ok {
		return fmt.Errorf("变量'%s'重新定义", name)
	}

	last[name] = twq
	return nil
}

func (ta *typeAnalyzer) markType(node ast.Node) {
	var err error
	switch n := node.(type) {
	case *ast.SimpleType:
		err = ta.simpleType(n)
	case *ast.SubType:
		err = ta.subType(n)
	case *ast.GenericType:
		err = ta.genericType(n)
	case *ast.BinaryExpr:
		err = ta.binaryExpr(n)
	case *ast.UnaryExpr:
		err = ta.unaryExpr(n)
	case *ast.AttrExpr:
		err = ta.attrExpr(n)
	case *ast.KwArg:
		err = ta.kwArg(n)
	case *ast.InstantiationExpr:
		err = ta.instantiationExpr(n)
	case *ast.CallExpr:
		err = ta.callExpr(n)
	case *ast.HRefExpr:
		err = ta.hRefExpr(n)
	case *ast.Identifier:
		err = ta.identifier(n)
	case *ast.StringLiteral:
		err = ta.stringLiteral(n)
	case *ast.IntLiteral:
		err = ta.intLiteral(n)
	case *ast.FloatLiteral:
		err = ta.floatLiteral(n)
	case *ast.ColorLiteral:
		err = ta.colorLiteral(n)
	case *ast.BoolLiteral:
		err = ta.boolLiteral(n)
	case *ast.TupleExpr:
		err = ta.tupleExpr(n)
	case *ast.TernaryExpr:
		err = ta.ternaryExpr(n)
	case *ast.ExprStmt:
		err = ta.exprStmt(n)
	case *ast.VarDeclStmt:
		err = ta.varDeclStmt(n)
	case *ast.TupleDeclStmt:
		err = ta.tupleDeclStmt(n)
	case *ast.ReassignStmt:
		err = ta.reassignStmt(n)
	case *ast.IfStmt:
		err = ta.ifStmt(n)
	case *ast.CaseClause:
		err = ta.caseClause(n)
	case *ast.SwitchStmt:
		err = ta.switchStmt(n)
	case *ast.WhileStmt:
		err = ta.whileStmt(n)
	case *ast.ForStmt:
		err = ta.forStmt(n)
	case *ast.ForInStmt:
		err = ta.forInStmt(n)
	case *ast.BreakStmt:
		err = ta.breakStmt(n)
	case *ast.ContinueStmt:
		err = ta.continueStmt(n)
	case *ast.ParamDecl:
		err = ta.paramDecl(n)
	case *ast.FuncDeclStmt:
		err = ta.funcDeclStmt(n)
	case *ast.MemberDecl:
		err = ta.memberDecl(n)
	case *ast.TypeDeclStmt:
		err = ta.typeDeclStmt(n)
	case *ast.ImportStmt:
		err = ta.importStmt(n)
	case *ast.Suite:
		err = ta.suite(n)
	case *ast.Quote:
		err = ta.quote(n)
	}

	if err != nil {
		ta.errors = append(ta.errors, err)
	}
}

func (ta *typeAnalyzer) simpleType(node *ast.SimpleType) error {
	name := node.Name
	parent := node.Parent()

	switch parent.(type) {
	case *ast.SubType, *ast.GenericType:
		if node.PathAttribute() == "Name" {
			t, err := ta.namespace.FindNamespace(name)
			if err != nil {
				return err
			}
			node.MarkNodeType(base.NSTypeWrap(*t))
			return nil
		}
	}

	t, err := ta.namespace.FindType(name)
	if err != nil {
		return err
	}

	node.MarkNodeType(t)
	return nil
}

func (ta *typeAnalyzer) subType(node *ast.SubType) error {
	ta.markType(node.Name)
	pt := node.Name.NodeType()
	smWrapper, ok := pt.(base.NSType)
	if !ok {
		return fmt.Errorf("'%s' is not a namespace", node.Name)
	}

	sm := smWrapper.Namespace
	t, err := sm.FindType(node.Member)
	if err != nil {
		return err
	}

	node.MarkNodeType(t)
	return nil
}

func (ta *typeAnalyzer) genericType(node *ast.GenericType) error {
	ta.markType(node.Name)
	args := []types.Type{}
	for _, arg := range node.Args {
		ta.markType(arg)
		args = append(args, arg.NodeType())
	}
	pt := node.Name.NodeType()
	ctor, ok := pt.(types.TypeOrCtor)
	if !ok {
		return fmt.Errorf("'%s' is not a type constructor", node.Name)
	}
	t, err := ctor.Ctor(args)
	if err != nil {
		return err
	}
	node.MarkNodeType(t)
	return nil
}

func (ta *typeAnalyzer) binaryExpr(node *ast.BinaryExpr) error {
	ta.markType(node.Left)
	ta.markType(node.Right)

	bop, ok := builtins.BinaryOperators[node.Op]
	if !ok {
		return fmt.Errorf("unknown operator '%s'", node.Op)
	}

	t, err := bop.Validate(node.Left.NodeType(), node.Right.NodeType())
	if err != nil {
		return err
	}

	node.MarkNodeType(t)
	return nil
}

func (ta *typeAnalyzer) unaryExpr(node *ast.UnaryExpr) error {
	ta.markType(node.Expr)

	uop, ok := builtins.UnaryOperators[node.Op]
	if !ok {
		return fmt.Errorf("unknown operator '%s'", node.Op)
	}

	t, err := uop.Validate(node.Expr.NodeType())
	if err != nil {
		return err
	}

	node.MarkNodeType(t)
	return nil
}

func (ta *typeAnalyzer) attrExpr(node *ast.AttrExpr) error {
	ta.markType(node.Target)
	p := node.Target.NodeType()
	switch p.Kind() {
	case types.StructKind:
		t := p.FieldByName(node.Name)
		if t == nil {
			return fmt.Errorf("'%s' has no attribute '%s'", node.Target, node.Name)
		}
		node.MarkNodeType(t.Type)
	case types.NamespaceKind:
		mw, ok := p.(base.NSType)
		if !ok {
			return fmt.Errorf("'%s' is not a namespace", mw)
		}
		m := mw.Namespace
		t, err := m.FindVariableType(node.Name)
		if err != nil {
			return err
		}
		node.MarkNodeType(t)
	default:
		// lookup method
		method, err := ta.namespace.FindMethod(node.Name, p)
		if err != nil {
			return err
		}
		node.MarkNodeType(types.CallableTypeWrap(method))
	}
	return nil
}

func (ta *typeAnalyzer) kwArg(node *ast.KwArg) error {
	ta.markType(node.Value)
	return nil
}

func (ta *typeAnalyzer) instantiationExpr(node *ast.InstantiationExpr) error {
	ta.markType(node.Template)
	return nil
}

func (ta *typeAnalyzer) callExpr(node *ast.CallExpr) error {
	ta.markType(node.Func)
	funcType := node.Func.NodeType()
	if funcType.Kind() != types.CallableKind {
		return fmt.Errorf("'%s' is not a callable", node.Func)
	}

	funcType = types.Peel(funcType)
	fn, ok := funcType.(types.CallableType)
	if !ok {
		return fmt.Errorf("'%s' cannot be casted into CallableType..., but why?", funcType.String())
	}
	argTypes := []types.Type{}
	kwArgTypes := map[string]types.Type{}
	for _, a := range node.Args {
		ta.markType(a)
		if kw, ok := a.(*ast.KwArg); ok {
			kwArgTypes[kw.Name] = a.NodeType()
		} else {
			argTypes = append(argTypes, a.NodeType())
		}
	}
	res, err := fn.Callable.Dispatch(argTypes, kwArgTypes)
	if err != nil {
		return err
	}

	node.MarkNodeType(res)
	return nil
}

func (ta *typeAnalyzer) hRefExpr(node *ast.HRefExpr) error {
	ta.markType(node.Series)
	node.MarkNodeType(types.TypeWithQualifier{
		Type:      node.Series.NodeType(),
		Qualifier: types.Series,
	})
	return nil
}

func (ta *typeAnalyzer) identifier(node *ast.Identifier) error {
	typ, err := ta.lookupIdentifier(node.Name)
	if err != nil {
		return err
	}
	node.MarkNodeType(typ)
	return nil
}

func (ta *typeAnalyzer) stringLiteral(node *ast.StringLiteral) error {
	node.MarkNodeType(types.String)
	return nil
}

func (ta *typeAnalyzer) intLiteral(node *ast.IntLiteral) error {
	node.MarkNodeType(types.Int)
	return nil
}

func (ta *typeAnalyzer) floatLiteral(node *ast.FloatLiteral) error {
	node.MarkNodeType(types.Float)
	return nil
}

func (ta *typeAnalyzer) colorLiteral(node *ast.ColorLiteral) error {
	node.MarkNodeType(types.Color)
	return nil
}

func (ta *typeAnalyzer) boolLiteral(node *ast.BoolLiteral) error {
	node.MarkNodeType(types.Bool)
	return nil
}

func (ta *typeAnalyzer) tupleExpr(node *ast.TupleExpr) error {
	items := []types.Type{}
	for _, item := range node.Items {
		ta.markType(item)
		items = append(items, item.NodeType())
	}

	node.MarkNodeType(types.TupleOf(items))
	return nil
}

func (ta *typeAnalyzer) ternaryExpr(node *ast.TernaryExpr) error {
	ta.markType(node.Test)
	ta.markType(node.True)
	ta.markType(node.False)
	trueType := node.True.NodeType()
	falseType := node.False.NodeType()
	if types.Equal(trueType, falseType) {
		node.MarkNodeType(trueType)
		return nil
	}

	if types.CanDoImplicitConversion(trueType, falseType) {
		node.MarkNodeType(falseType)
		return nil
	}

	if types.CanDoImplicitConversion(falseType, trueType) {
		node.MarkNodeType(trueType)
		return nil
	}

	return fmt.Errorf("type mismatch in ternary expression: %s and %s", trueType.String(), falseType.String())
}

func (ta *typeAnalyzer) exprStmt(node *ast.ExprStmt) error {
	ta.markType(node.Expr)
	node.MarkNodeType(node.Expr.NodeType())
	return nil
}

func (ta *typeAnalyzer) varDeclStmt(node *ast.VarDeclStmt) error {
	var formalType types.Type = nil
	if node.Type != nil {
		ta.markType(node.Type)
		formalType = node.Type.NodeType()
	}
	ta.markType(node.Initial)
	initType := node.Initial.NodeType()
	if formalType == nil {
		if !initType.Kind().IsNormal() {
			return fmt.Errorf("cannot infer type of '%s', init stmt type is '%s'", node.Name, initType.String())
		}

		formalType = initType
	} else if !types.Equal(formalType, initType) && !types.CanDoImplicitConversion(initType, formalType) {
		return fmt.Errorf("type mismatch: '%s' expect a '%s' value, but got '%s'", node.Name, formalType.String(), initType.String())
	}

	qualifier := types.NoQualifier
	if node.Qualifier != nil {
		switch *node.Qualifier {
		case "const":
			qualifier = types.Const
		case "simple":
			qualifier = types.Simple
		case "series":
			qualifier = types.Series
		}
	}

	if qualifier == types.NoQualifier && initType.QualifierKind() != types.NoQualifier {
		// qualifier 'input' are from input function series call
		qualifier = initType.QualifierKind()
	}

	twq := types.TypeWithQualifier{
		Type:      formalType,
		Qualifier: qualifier,
	}

	node.MarkNodeType(twq)

	if err := ta.registerVariable(node.Name, twq); err != nil {
		return err
	}

	return nil
}

func (ta *typeAnalyzer) tupleDeclStmt(node *ast.TupleDeclStmt) error {
	ta.markType(node.Initial)
	if node.Initial.NodeType().Kind() != types.TupleKind {
		return fmt.Errorf("%s返回的不是一个元组，因此无法对元组赋值", node.Initial)
	}

	varsCount := node.Initial.NodeType().Count()
	if varsCount != len(node.Variables) {
		return fmt.Errorf("尝试将%d个元素的元组赋值给%d个变量", varsCount, len(node.Variables))
	}

	for i, v := range node.Variables {
		if err := ta.registerVariable(v, types.TypeWithQualifier{
			Type:      node.Initial.NodeType().Item(i),
			Qualifier: types.NoQualifier,
		}); err != nil {
			return err
		}
	}

	node.MarkNodeType(node.Initial.NodeType())

	return nil
}

func (ta *typeAnalyzer) reassignStmt(node *ast.ReassignStmt) error {
	ta.markType(node.Target)
	if !node.Target.NodeType().Kind().IsValid() {
		return nil
	}
	ta.markType(node.Value)
	node.MarkNodeType(node.Target.NodeType())
	return nil
}

func (ta *typeAnalyzer) ifStmt(node *ast.IfStmt) error {
	ta.markType(node.Test)

	if node.Test.NodeType().Kind() != types.BoolKind && !types.CanDoImplicitConversion(node.Test.NodeType(), types.Bool) {
		return fmt.Errorf("if语句的条件表达式需为bool类型，而不是%s", node.Test.NodeType().String())
	}

	var trueType types.Type = types.Void
	if node.True != nil {
		ta.markType(node.True)
		trueType = node.True.NodeType()
	}

	var falseType types.Type = types.Void
	if node.False != nil {
		ta.markType(node.False)
		falseType = node.False.NodeType()
	}

	node.MarkNodeType(types.Union(trueType, falseType))
	return nil
}

func (ta *typeAnalyzer) caseClause(node *ast.CaseClause) error {
	s, ok := node.Parent().(*ast.SwitchStmt)
	if !ok {
		return fmt.Errorf("case子语句必须在switch语句中使用")
	}

	if s.Target == nil {
		ta.markType(node.Cond)
		if node.Cond.NodeType().Kind() != types.BoolKind && !types.CanDoImplicitConversion(node.Cond.NodeType(), types.Bool) {
			return fmt.Errorf("如果不提供switch的对象，则case子句的条件必须为bool类型，而不是%s", node.Cond.NodeType().String())
		}
	} else {
		ta.markType(node.Cond)
		if !types.Equal(node.Cond.NodeType(), s.Target.NodeType()) && !types.CanDoImplicitConversion(node.Cond.NodeType(), s.Target.NodeType()) {
			return fmt.Errorf("case子句的条件类型是%s，而需要的类型是%s", node.Cond.NodeType().String(), s.Target.NodeType().String())
		}
	}

	ta.markType(node.Body)
	node.MarkNodeType(node.Body.NodeType())
	return nil
}

func (ta *typeAnalyzer) switchStmt(node *ast.SwitchStmt) error {
	ta.markType(node.Target)

	ts := []types.Type{}

	for _, c := range node.Cases {
		ta.markType(c)
		ts = append(ts, c.NodeType())
	}

	ta.markType(node.Default)
	ts = append(ts, node.Default.NodeType())

	node.MarkNodeType(types.UnionOf(ts))
	return nil
}

func (ta *typeAnalyzer) whileStmt(node *ast.WhileStmt) error {
	ta.markType(node.Test)

	if node.Test.NodeType().Kind() != types.BoolKind && !types.CanDoImplicitConversion(node.Test.NodeType(), types.Bool) {
		return fmt.Errorf("while语句的条件表达式需为bool类型，而不是%s", node.Test.NodeType().String())
	}

	ta.markType(node.Body)
	node.MarkNodeType(node.Body.NodeType())
	return nil
}

func (ta *typeAnalyzer) forStmt(node *ast.ForStmt) error {
	ta.markType(node.Init)
	if !types.Equal(node.Init.NodeType(), types.Int) && !types.Equal(node.Init.NodeType(), types.Float) {
		return fmt.Errorf("for循环的初值只能是整数或浮点数")
	}

	if node.Step != nil {
		ta.markType(node.Step)
		if !types.Equal(node.Step.NodeType(), types.Int) && !types.Equal(node.Step.NodeType(), types.Float) {
			return fmt.Errorf("for循环的步进值只能是整数或浮点数")
		}
	}

	ta.markType(node.Final)
	if !types.Equal(node.Final.NodeType(), types.Int) && !types.Equal(node.Final.NodeType(), types.Float) {
		return fmt.Errorf("for循环的终值只能是整数或浮点数")
	}

	ta.markType(node.Body)
	node.MarkNodeType(node.Body.NodeType())
	return nil
}

func (ta *typeAnalyzer) forInStmt(node *ast.ForInStmt) error {
	return nil
}

func (ta *typeAnalyzer) breakStmt(node *ast.BreakStmt) error {
	node.MarkNodeType(types.Void)
	return nil
}

func (ta *typeAnalyzer) continueStmt(node *ast.ContinueStmt) error {
	node.MarkNodeType(types.Void)
	return nil
}

func (ta *typeAnalyzer) paramDecl(node *ast.ParamDecl) error {
	qualifier := types.NoQualifier
	if node.Qualifier != nil {
		switch *node.Qualifier {
		case "const":
			qualifier = types.Const
		case "simple":
			qualifier = types.Simple
		case "series":
			qualifier = types.Series
		}
	}

	var formalType types.Type = types.Uncertain
	if node.Type != nil {
		ta.markType(node.Type)
		formalType = node.Type.NodeType()
	}

	if node.Default != nil {
		ta.markType(node.Default)
		initType := node.Default.NodeType()

		if formalType.Kind() == types.UncertainKind {
			formalType = initType
		} else {
			if !types.Equal(formalType, initType) && !types.CanDoImplicitConversion(initType, formalType) {
				return fmt.Errorf("参数%s的类型为%s，但其初始值的类型却是%s", node.Name, formalType, initType)
			}
		}
	}

	node.MarkNodeType(types.TypeWithQualifier{
		Type: formalType,
		Qualifier: qualifier,
	})
	return nil
}

func (ta *typeAnalyzer) funcDeclStmt(node *ast.FuncDeclStmt) error {
	ins := []types.TypeWithName{}
	for _, p := range node.Params {
		ta.markType(p)
		ins = append(ins, types.TypeWithName{
			Name: p.Name,
			Type: p.NodeType(),
			Optional: p.Default != nil,
		})
	}

	ta.markType(node.Body)
	fnType := types.FunctionOf(ins, node.Body.NodeType())

	node.MarkNodeType(fnType)
	return nil
}

func (ta *typeAnalyzer) memberDecl(node *ast.MemberDecl) error {
	ta.markType(node.Type)
	formalType := node.Type.NodeType()
	if formalType == nil {
		return nil
	}

	if p, ok := node.Parent().(*ast.TypeDeclStmt); p != nil && ok {
		return fmt.Errorf("成员变量定义语句只能在定义类型的上下文中使用")
	} else {
		if node.Default != nil {
			switch node.Default.(type) {
			case *ast.IntLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.ColorLiteral, *ast.BoolLiteral, *ast.Identifier:
				ta.markType(node.Default)
				if !types.Equal(node.Default.NodeType(), formalType) && types.CanDoImplicitConversion(node.Default.NodeType(), formalType) {
					return fmt.Errorf("自定义类型%s的成员变量%s类型为%s，但其初始值的类型却是%s", p.Name, node.Name, formalType.String(), node.Default.NodeType().String())
				}
			default:
				return fmt.Errorf("只能使用内置变量或字面量声明成员变量的默认值")
			}
		}
		node.MarkNodeType(formalType)
		return nil
	}
}

func (ta *typeAnalyzer) typeDeclStmt(node *ast.TypeDeclStmt) error {
	if _, err := ta.userNS.FindType(node.Name); err == nil {
		return fmt.Errorf("类型%s被重定义", node.Name)
	}

	fields := []types.TypeWithName{}
	for _, m := range node.Members {
		ta.markType(m)
		fields = append(fields, types.TypeWithName{
			Name:     m.Name,
			Type:     m.NodeType(),
			Optional: m.Default != nil,
		})
	}
	st := types.StructOf(fields)
	node.MarkNodeType(st)

	ta.userNS.Types[node.Name] = types.NewTocType(st)
	return nil
}

func (ta *typeAnalyzer) importStmt(node *ast.ImportStmt) error {
	node.MarkNodeType(types.Void)
	return nil
}

func (ta *typeAnalyzer) suite(node *ast.Suite) error {
	ta.enterScope()
	for _, stmt := range node.Body {
		ta.markType(stmt)
	}
	ta.exitScope()
	node.MarkNodeType(node.Body[len(node.Body)-1].NodeType())
	return nil
}

func (ta *typeAnalyzer) quote(node *ast.Quote) error {
	ta.markType(node.Content)
	node.MarkNodeType(node.Content.NodeType())
	return nil
}
