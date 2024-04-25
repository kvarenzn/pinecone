package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kvarenzn/pinecone/ast"
	"github.com/kvarenzn/pinecone/tokenizer"
)

func parseNumber(token tokenizer.Token) ast.Node {
	if token.Type != tokenizer.NUMBER {
		return nil
	}
	lexeme := token.Lexeme
	if strings.Contains(lexeme, ".") {
		num, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			return nil
		}
		return ast.WithRange(&ast.FloatLiteral{
			Value: num,
		}, token.Begin, token.End)
	}
	num, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		return nil
	}
	return ast.WithRange(&ast.IntLiteral{
		Value: num,
	}, token.Begin, token.End)
}

func parseString(token tokenizer.Token) ast.Node {
	if token.Type != tokenizer.STRING {
		return nil
	}
	lexeme := token.Lexeme
	lexeme = strings.ReplaceAll(lexeme, "\n", "")
	if lexeme[0] == '\'' {
		lexeme = lexeme[1 : len(lexeme)-1]
		lexeme = strings.ReplaceAll(lexeme, "\\'", "'")
		lexeme = strings.ReplaceAll(lexeme, "\"", "\\\"")
		lexeme = `"` + lexeme + `"`
	}
	str, err := strconv.Unquote(lexeme)
	if err != nil {
		return nil
	}

	return ast.WithRange(&ast.StringLiteral{
		Value: str,
	}, token.Begin, token.End)
}

func parseColor(token tokenizer.Token) ast.Node {
	if token.Type != tokenizer.COLOR {
		return nil
	}

	lexeme := token.Lexeme
	if lexeme[0] != '#' {
		return nil
	}

	num, err := strconv.ParseInt(lexeme[1:], 16, 64)
	if err != nil {
		return nil
	}
	var red, green, blue, transparent float64 = 0, 0, 0, 0
	switch len(lexeme) {
	case 4: // #RGB
		red = float64((num>>8)&0xf) * 0x11
		green = float64((num>>4)&0xf) * 0x11
		blue = float64((num>>0)&0xf) * 0x11
		transparent = 0.0
	case 5: // #RGBA
		red = float64((num>>12)&0xf) * 0x11
		green = float64((num>>8)&0xf) * 0x11
		blue = float64((num>>4)&0xf) * 0x11
		transparent = 100 - float64((num>>0)&0xf)/0xf*100
	case 7: // #RRGGBB
		red = float64((num >> 16) & 0xff)
		green = float64((num >> 8) & 0xff)
		blue = float64((num >> 0) & 0xff)
		transparent = 0
	case 9: // #RRGGBBAA
		red = float64((num>>24)&0xff) / 0xff
		green = float64((num>>16)&0xff) / 0xff
		blue = float64((num>>8)&0xff) / 0xff
		transparent = 100 - float64((num>>0)&0xff)/0xff*100
	default:
		return nil
	}
	return ast.WithRange(&ast.ColorLiteral{
		R: red,
		G: green,
		B: blue,
		T: transparent,
	}, token.Begin, token.End)
}

func pickLexeme(token *tokenizer.Token) *string {
	if token == nil {
		return nil
	}

	return &token.Lexeme
}

type ParseError struct {
	Row int
	Col int
	Msg string
}

type parser struct {
	tokens  []tokenizer.Token
	current int
	errors  []ParseError
}

func (p parser) eof() bool {
	return p.current >= len(p.tokens)
}

func (p parser) peek(n int) *tokenizer.Token {
	if p.current+n >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.current+n]
}

func (p parser) peekType(n int) tokenizer.TokenType {
	token := p.peek(n)
	if token == nil {
		return tokenizer.UNKNOWN
	}
	return token.Type
}

func (p parser) peekLexeme() string {
	token := p.peek(0)
	if token == nil {
		return "\"\""
	}

	return fmt.Sprintf("%q", token.Lexeme)
}

func (p parser) tell() int {
	return p.current
}

func (p *parser) seek(pos int) {
	p.current = pos
}

func (p *parser) consume(tts ...tokenizer.TokenType) *tokenizer.Token {
	token := p.peek(0)
	if token == nil {
		return nil
	}
	if len(tts) > 0 {
		found := false
		for _, tt := range tts {
			if tt == token.Type {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	p.current++
	fmt.Println(token)
	return token
}

func (p *parser) seekToType(tt tokenizer.TokenType) {
	for p.peekType(0) != tt && !p.eof() {
		p.consume()
	}
}

func (p *parser) hasTokenBeforeNewLine(tt tokenizer.TokenType) bool {
	offset := 1
	for {
		t := p.peekType(offset)
		if t.In(tokenizer.UNKNOWN, tokenizer.NEWLINE, tokenizer.INDENT, tokenizer.DEDENT) {
			break
		}
		if t == tt {
			return true
		}
		offset++
	}
	return false
}

func (p *parser) next() {
	p.current++
}

func (p *parser) prev() {
	p.current--
}

func (p *parser) error(format string, args ...any) {
	token := p.peek(0)
	msg := fmt.Sprintf(format, args...)
	if token == nil {
		p.errors = append(p.errors, ParseError{
			Row: -1,
			Col: -1,
			Msg: msg,
		})
	} else {
		p.errors = append(p.errors, ParseError{
			Row: token.Begin.Row,
			Col: token.Begin.Column,
			Msg: msg,
		})
	}
}

func (p *parser) getIdentifier() *tokenizer.Token {
	token := p.peek(0)
	if token == nil || token.Type != tokenizer.IDENTIFIER && !token.IsSoftKeyword() {
		return nil
	}
	p.consume()
	return token
}

func (p *parser) parseType(silent bool) ast.Node {
	name := p.getIdentifier()
	if name == nil {
		if !silent {
			p.error("Expect a type, but got %s", p.peekLexeme())
		}
		return nil
	}
	var t ast.Node = &ast.SimpleType{
		Name: name.Lexeme,
	}

	for {
		token := p.consume(tokenizer.LEFT_ANG_BRACKET, tokenizer.DOT, tokenizer.LEFT_SQ_BRACKET)
		if token == nil {
			break
		}
		switch token.Type {
		case tokenizer.LEFT_ANG_BRACKET:
			types := p.parseTypeArgList(silent)
			if types == nil {
				return nil
			}
			rang := p.consume(tokenizer.RIGHT_ANG_BRACKET)
			if rang == nil {
				if !silent {
					p.error(`Expect ">" to match "<", but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = ast.WithRange(&ast.GenericType{
				Name: t,
				Args: types,
			}, t.Begin(), rang.End)
		case tokenizer.DOT:
			name := p.getIdentifier()
			if name == nil {
				if !silent {
					p.error(`Expect an identifier, but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = ast.WithRange(&ast.SubType{
				Name:   t,
				Member: name.Lexeme,
			}, t.Begin(), name.End)
		case tokenizer.LEFT_SQ_BRACKET:
			// type[] => array<type>
			rsq := p.consume(tokenizer.RIGHT_SQ_BRACKET)
			if rsq == nil {
				if !silent {
					p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = ast.WithRange(&ast.GenericType{
				Name: &ast.SimpleType{
					Name: "array",
				},
				Args: []ast.Node{t},
			}, t.Begin(), rsq.End)
		}
	}

	return t
}

func (p *parser) parseTypeArgList(silent bool) []ast.Node {
	typeArgs := []ast.Node{}
	for {
		if p.peekType(0).In(tokenizer.UNKNOWN, tokenizer.RIGHT_ANG_BRACKET, tokenizer.NEWLINE) {
			break
		}
		typeArg := p.parseType(silent)
		if typeArg == nil {
			return nil
		}
		typeArgs = append(typeArgs, typeArg)
		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}
	return typeArgs
}

func (p *parser) parseTupleAtom(silent bool) ast.Node {
	lsq := p.consume(tokenizer.LEFT_SQ_BRACKET)
	if lsq == nil {
		if !silent {
			p.error(`Expect "[", but got %s`, p.peekLexeme())
		}
		return nil
	}
	items := &ast.TupleExpr{
		Items: []ast.Node{},
	}

	for {
		if p.peekType(0).In(tokenizer.UNKNOWN, tokenizer.RIGHT_SQ_BRACKET, tokenizer.NEWLINE) {
			break
		}
		item := p.parseTestExpr(silent)
		if item == nil {
			return nil
		}

		items.Items = append(items.Items, item)
		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}

	rsq := p.consume(tokenizer.RIGHT_SQ_BRACKET)
	if rsq == nil {
		if !silent {
			p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
		}
		return nil
	}

	return ast.WithRange(items, lsq.Begin, rsq.End)
}

func (p *parser) parseParenExpr(silent bool) ast.Node {
	lparen := p.consume(tokenizer.LEFT_PAREN)
	if lparen == nil {
		if !silent {
			p.error(`Expect "(", but got %s`, p.peekLexeme())
		}
		return nil
	}

	expr := p.parseTestExpr(silent)
	if expr == nil {
		return nil
	}

	rparen := p.consume(tokenizer.RIGHT_PAREN)
	if rparen == nil {
		if !silent {
			p.error(`Expect ")" to match "(", but got %s`, p.peekLexeme())
		}
	}

	return ast.WithRange(expr, lparen.Begin, rparen.End)
}

func (p *parser) parseAtom(silent bool) ast.Node {
	token := p.peek(0)
	if token == nil {
		if !silent {
			p.error(`Unexpected EOF, file might be truncated`)
		}
	}

	switch token.Type {
	case tokenizer.NUMBER:
		num := parseNumber(*token)
		if num == nil {
			p.error(`Invalid number %s`, p.peekLexeme())
			return nil
		}
		p.consume(tokenizer.NUMBER)
		return num
	case tokenizer.STRING:
		str := parseString(*token)
		if str == nil {
			p.error(`Invalid string literal %s`, p.peekLexeme())
			return nil
		}
		p.consume(tokenizer.STRING)
		return str
	case tokenizer.COLOR:
		color := parseColor(*token)
		if color == nil {
			p.error(`Invalid color literal %s`, p.peekLexeme())
			return nil
		}
		p.consume(tokenizer.COLOR)
		return nil
	case tokenizer.TRUE:
		p.consume(tokenizer.TRUE)
		return ast.WithRange(&ast.TrueExpr{}, token.Begin, token.End)
	case tokenizer.FALSE:
		p.consume(tokenizer.FALSE)
		return ast.WithRange(&ast.FalseExpr{}, token.Begin, token.End)
	case tokenizer.LEFT_PAREN:
		return p.parseParenExpr(silent)
	case tokenizer.LEFT_SQ_BRACKET:
		return p.parseTupleAtom(silent)
	default:
		id := p.getIdentifier()
		if id != nil {
			return ast.WithRange(&ast.Identifier{
				Name: id.Lexeme,
			}, id.Begin, id.End)
		}
	}

	if !silent {
		p.error(`Expect an identifier, a number, string, color, bool, paren or tuple, but got %s`, p.peekLexeme())
	}

	return nil
}

func (p *parser) parseArgument(silent bool) ast.Node {
	value := p.parseTestExpr(silent)
	if value == nil {
		return nil
	}
	if p.consume(tokenizer.EQUAL) == nil {
		return value
	}
	name := value
	value = p.parseTestExpr(silent)
	if value == nil {
		return nil
	}
	if n, ok := name.(*ast.Identifier); ok {
		return ast.WithRange(&ast.KwArg{
			Name:  n.Name,
			Value: value,
		}, n.Begin(), value.End())
	} else {
		return nil
	}
}

func (p *parser) parseArgList(silent bool) []ast.Node {
	args := []ast.Node{}
	for {
		token := p.peek(0)
		if token == nil || token.Type.In(tokenizer.RIGHT_PAREN, tokenizer.NEWLINE) {
			break
		}

		arg := p.parseArgument(silent)
		if arg == nil {
			return nil
		}
		args = append(args, arg)
		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}
	return args
}

func (p *parser) parseAtomExpr(silent bool) ast.Node {
	atom := p.parseAtom(silent)
	if atom == nil {
		return nil
	}
	for {
		if p.consume(tokenizer.LEFT_PAREN) != nil {
			args := p.parseArgList(silent)
			if args == nil {
				return nil
			}
			rparen := p.consume(tokenizer.RIGHT_PAREN)
			if rparen == nil {
				if !silent {
					p.error(`Expect ")" to match "(", but got %s`, p.peekLexeme())
				}
				return nil
			}
			atom = ast.WithRange(&ast.CallExpr{
				Func: atom,
				Args: args,
			}, atom.Begin(), rparen.End)
		} else if p.consume(tokenizer.LEFT_SQ_BRACKET) != nil {
			offset := p.parseTestExpr(silent)
			if offset == nil {
				return nil
			}
			if p.consume(tokenizer.RIGHT_SQ_BRACKET) == nil {
				if !silent {
					p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
				}
				return nil
			}
		} else if p.consume(tokenizer.LEFT_ANG_BRACKET) != nil {
			begin := p.tell()
			typeArgs := p.parseTypeArgList(true)
			if typeArgs == nil || p.peekType(0) != tokenizer.RIGHT_ANG_BRACKET || p.peekType(1) != tokenizer.LEFT_PAREN {
				p.seek(begin - 1)
				return atom
			}
			rang := p.consume(tokenizer.RIGHT_ANG_BRACKET)
			atom = ast.WithRange(&ast.InstantiationExpr{
				Template: atom,
				TypeArgs: typeArgs,
			}, atom.Begin(), rang.End)
		} else if p.consume(tokenizer.DOT) != nil {
			member := p.getIdentifier()
			if member == nil {
				if !silent {
					p.error(`Expect an identifier as member name, but got %s`, p.peekLexeme())
				}
				return nil
			}
			atom = ast.WithRange(&ast.AttrExpr{
				Target: atom,
				Name:   member.Lexeme,
			}, atom.Begin(), member.End)
		} else {
			break
		}
	}

	return atom
}

func (p *parser) parseArithmeticFactor(silent bool) ast.Node {
	op := p.consume(tokenizer.PLUS, tokenizer.MINUS)
	if op != nil {
		expr := p.parseArithmeticFactor(silent)
		if expr == nil {
			return nil
		}
		return ast.WithRange(&ast.UnaryExpr{
			Op:   op.Lexeme,
			Expr: expr,
		}, op.Begin, expr.End())
	}
	expr := p.parseAtomExpr(silent)
	if expr == nil {
		return nil
	}
	return expr
}

func (p *parser) parseArithmeticTerm(silent bool) ast.Node {
	left := p.parseArithmeticFactor(silent)
	if left == nil {
		return nil
	}
	for {
		op := p.consume(tokenizer.STAR, tokenizer.SLASH, tokenizer.PERCENT)
		if op == nil {
			break
		}
		right := p.parseArithmeticFactor(silent)
		if right == nil {
			return nil
		}
		left = ast.WithRange(&ast.BinaryExpr{
			Left:  left,
			Op:    op.Lexeme,
			Right: right,
		}, left.Begin(), right.End())
	}
	return left
}

func (p *parser) parseArithmeticExpr(silent bool) ast.Node {
	left := p.parseArithmeticTerm(silent)
	if left == nil {
		return nil
	}
	for {
		op := p.consume(tokenizer.PLUS, tokenizer.MINUS)
		if op == nil {
			break
		}
		right := p.parseArithmeticTerm(silent)
		if right == nil {
			return nil
		}
		left = ast.WithRange(&ast.BinaryExpr{
			Left:  left,
			Op:    op.Lexeme,
			Right: right,
		}, left.Begin(), right.End())
	}
	return left
}

func (p *parser) parseComparison(silent bool) ast.Node {
	left := p.parseArithmeticExpr(silent)
	if left == nil {
		return nil
	}
	for {
		op := p.consume(tokenizer.LEFT_ANG_BRACKET, tokenizer.RIGHT_ANG_BRACKET, tokenizer.LESS_EQUAL, tokenizer.GREATER_EQUAL, tokenizer.EQUAL_EQUAL, tokenizer.BANG_EQUAL)
		if op == nil {
			break
		}
		right := p.parseArithmeticExpr(silent)
		if right == nil {
			return nil
		}
		left = ast.WithRange(&ast.BinaryExpr{
			Left:  left,
			Op:    op.Lexeme,
			Right: right,
		}, left.Begin(), right.End())
	}
	return left
}

func (p *parser) parseNotTest(silent bool) ast.Node {
	not := p.consume(tokenizer.NOT)
	if not == nil {
		return p.parseComparison(silent)
	}
	notExpr := p.parseNotTest(silent)
	if notExpr == nil {
		return nil
	}
	return ast.WithRange(&ast.UnaryExpr{
		Op:   not.Lexeme,
		Expr: notExpr,
	}, not.Begin, notExpr.End())
}

func (p *parser) parseAndTest(silent bool) ast.Node {
	left := p.parseNotTest(silent)
	if left == nil {
		return nil
	}
	for {
		and := p.consume(tokenizer.AND)
		if and == nil {
			break
		}
		right := p.parseNotTest(silent)
		if right == nil {
			return nil
		}
		left = ast.WithRange(&ast.BinaryExpr{
			Left:  left,
			Op:    and.Lexeme,
			Right: right,
		}, left.Begin(), right.End())
	}
	return left
}

func (p *parser) parseOrTest(silent bool) ast.Node {
	left := p.parseAndTest(silent)
	if left == nil {
		return nil
	}
	for {
		or := p.consume(tokenizer.OR)
		if or == nil {
			break
		}
		right := p.parseAndTest(silent)
		if right == nil {
			return nil
		}
		left = ast.WithRange(&ast.BinaryExpr{
			Left:  left,
			Op:    or.Lexeme,
			Right: right,
		}, left.Begin(), right.End())
	}
	return left
}

func (p *parser) parseTestExpr(silent bool) ast.Node {
	test := p.parseOrTest(silent)
	if test == nil {
		return nil
	}
	if p.consume(tokenizer.QUESTION) == nil {
		return test
	}
	t := p.parseTestExpr(silent)
	if t == nil {
		return nil
	}
	if p.consume(tokenizer.COLON) == nil {
		if !silent {
			p.error(`Expect ":" to match "?", but got %s`, p.peekLexeme())
		}
		return nil
	}
	f := p.parseTestExpr(silent)
	if f == nil {
		return nil
	}
	return ast.WithRange(&ast.TernaryExpr{
		Test:  test,
		True:  t,
		False: f,
	}, test.Begin(), f.End())
}

func (p *parser) parseIfStmt() ast.Node {
	ifToken := p.consume(tokenizer.IF)
	if ifToken == nil {
		p.error(`Expect "if", but got %s`, p.peekLexeme())
		return nil
	}

	test := p.parseTestExpr(false)
	if test == nil {
		return nil
	}

	suite := p.parseSuite()
	if suite == nil {
		return nil
	}

	ifStmt := &ast.IfStmt{
		Test: test,
		True: suite,
	}

	ifStmt.SetBegin(ifToken.Begin)

	if p.consume(tokenizer.ELSE) != nil {
		f := p.parseSuite()
		if f == nil {
			return nil
		}
		ifStmt.False = f
		ifStmt.SetEnd(f.End())
	}

	return ifStmt
}

func (p *parser) parseWhileStmt() ast.Node {
	w := p.consume(tokenizer.WHILE)
	if w == nil {
		p.error(`Expect "while", but got %s`, p.peekLexeme())
		return nil
	}

	test := p.parseTestExpr(false)
	if test == nil {
		return nil
	}
	suite := p.parseSuite()
	if suite == nil {
		return nil
	}
	return ast.WithRange(&ast.WhileStmt{
		Test: test,
		Body: suite,
	}, w.Begin, suite.End())
}

func (p *parser) parseForStmt() ast.Node {
	f := p.consume(tokenizer.FOR)
	if f == nil {
		p.error(`Expect "for", but got %s`, p.peekLexeme())
		return nil
	}

	token := p.peek(0)
	if token == nil {
		p.error(`Uncomplete for statement`)
		return nil
	}

	if p.consume(tokenizer.LEFT_SQ_BRACKET) != nil {
		idx := p.getIdentifier()
		if idx == nil {
			p.error(`Expect an identifier as index variable, but got %s`, p.peekLexeme())
			return nil
		}
		if p.consume(tokenizer.COMMA) == nil {
			p.error(`Expect ",", but got %s`, p.peekLexeme())
			return nil
		}
		iter := p.getIdentifier()
		if iter == nil {
			p.error(`Expect an identifier as iterator variable, but got %s`, p.peekLexeme())
			return nil
		}
		if p.consume(tokenizer.RIGHT_SQ_BRACKET) == nil {
			p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
			return nil
		}
		if p.consume(tokenizer.IN) == nil {
			p.error(`Expect "in", but got %s`, p.peekLexeme())
			return nil
		}

		container := p.parseTestExpr(false)
		if container == nil {
			return nil
		}
		suite := p.parseSuite()
		if suite == nil {
			return nil
		}

		return ast.WithRange(&ast.ForInStmt{
			Index:     &idx.Lexeme,
			Iterator:  iter.Lexeme,
			Container: container,
			Body:      suite,
		}, f.Begin, suite.End())
	}

	counter := p.getIdentifier()
	if counter == nil {
		p.error(`Expect an identifier as loop variable, but got %s`, p.peekLexeme())
		return nil
	}

	if p.consume(tokenizer.EQUAL) != nil {
		init := p.parseTestExpr(false)
		if init == nil {
			return nil
		}
		if p.consume(tokenizer.TO) == nil {
			p.error(`Expect "to", but got %s`, p.peekLexeme())
			return nil
		}
		final := p.parseTestExpr(false)
		if final == nil {
			return nil
		}

		var step ast.Node = nil
		if p.consume(tokenizer.BY) != nil {
			step = p.parseTestExpr(false)
			if step == nil {
				return nil
			}
		}

		suite := p.parseSuite()
		if suite == nil {
			return nil
		}

		return ast.WithRange(&ast.ForStmt{
			Counter: counter.Lexeme,
			Init:    init,
			Step:    step,
			Final:   final,
			Body:    suite,
		}, f.Begin, suite.End())
	} else if p.consume(tokenizer.IN) != nil {
		container := p.parseTestExpr(false)
		if container == nil {
			return nil
		}

		suite := p.parseSuite()
		if suite == nil {
			return nil
		}

		return ast.WithRange(&ast.ForInStmt{
			Index:     nil,
			Iterator:  counter.Lexeme,
			Container: container,
			Body:      suite,
		}, f.Begin, suite.End())
	}

	p.error(`Expect "in" or "=", but got %s`, p.peekLexeme())
	return nil
}

func (p *parser) parseCaseClause() ast.Node {
	if p.consume(tokenizer.RIGHT_FAT_ARROW) != nil {
		return p.parseSuite()
	}

	cond := p.parseTestExpr(false)
	if cond == nil {
		return nil
	}

	if p.consume(tokenizer.RIGHT_FAT_ARROW) == nil {
		p.error(`Expect "=>", but got %s`, p.peekLexeme())
		return nil
	}

	body := p.parseSuite()
	if body == nil {
		return nil
	}

	return ast.WithRange(&ast.CaseClause{
		Cond: cond,
		Body: body,
	}, cond.Begin(), body.End())
}

func (p *parser) parseSwitchStmt() ast.Node {
	sw := p.consume(tokenizer.SWITCH)
	if sw == nil {
		p.error(`Expect "switch", but got %s`, p.peekLexeme())
		return nil
	}

	var target ast.Node = nil

	if p.peekType(0) != tokenizer.INDENT {
		target = p.parseTestExpr(false)
		if target == nil {
			return nil
		}
	}

	if p.consume(tokenizer.INDENT) == nil {
		p.error(`Expect indent, but got %s`, p.peekLexeme())
		return nil
	}

	switchStmt := &ast.SwitchStmt{
		Target:  target,
		Cases:   []*ast.CaseClause{},
		Default: nil,
	}
	switchStmt.SetBegin(sw.Begin)

	for {
		token := p.peek(0)
		if token == nil {
			p.error(`Uncomplete switch statement`)
			return nil
		}

		if token.Type == tokenizer.DEDENT {
			p.consume(tokenizer.DEDENT)
			switchStmt.SetEnd(token.End)
			break
		}

		caseClause := p.parseCaseClause()
		if caseClause == nil {
			return nil
		}

		switch c := caseClause.(type) {
		case *ast.CaseClause:
			switchStmt.Cases = append(switchStmt.Cases, c)
		default:
			switchStmt.Default = c
		}
		p.consume(tokenizer.NEWLINE)
	}
	return switchStmt
}

func (p *parser) parseVarDeclStmt() ast.Node {
	declMode := p.consume(tokenizer.VARIP, tokenizer.VAR)
	qualifier := p.consume(tokenizer.SERIES, tokenizer.CONST, tokenizer.SIMPLE)

	if p.peekType(1) == tokenizer.EQUAL {
		name := p.getIdentifier()
		if name == nil {
			p.error(`Expect an identifier as variable name, but got %s`, p.peekLexeme())
			return nil
		}
		p.consume(tokenizer.EQUAL)
		init := p.parseStmt()
		if init == nil {
			return nil
		}
		begin := name.Begin
		if qualifier != nil {
			begin = qualifier.Begin
		}
		if declMode != nil {
			begin = declMode.Begin
		}
		return ast.WithRange(&ast.VarDeclStmt{
			DeclMode:  pickLexeme(declMode),
			Qualifier: pickLexeme(qualifier),
			Type:      nil,
			Name:      name.Lexeme,
			Initial:   init,
		}, begin, init.End())
	}

	t := p.parseType(false)
	if t == nil {
		return nil
	}

	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as variable name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.consume(tokenizer.EQUAL) == nil {
		p.error(`Expect "=" and a expression, but got %s`, p.peekLexeme())
		return nil
	}

	init := p.parseStmt()
	if init == nil {
		return nil
	}

	begin := t.Begin()
	if qualifier != nil {
		begin = qualifier.Begin
	}
	if declMode != nil {
		begin = declMode.Begin
	}
	return ast.WithRange(&ast.VarDeclStmt{
		DeclMode:  pickLexeme(declMode),
		Qualifier: pickLexeme(qualifier),
		Type:      t,
		Name:      name.Lexeme,
		Initial:   init,
	}, begin, init.End())
}

func (p *parser) parseIdentifierTuple(silent bool) []tokenizer.Token {
	if p.consume(tokenizer.LEFT_SQ_BRACKET) == nil {
		if !silent {
			p.error(`Expect "[", but got %s`, p.peekLexeme())
		}
		return nil
	}

	ids := []tokenizer.Token{}
	for {
		token := p.peek(0)
		if token == nil || token.Type == tokenizer.RIGHT_SQ_BRACKET || token.Type == tokenizer.NEWLINE {
			break
		}

		name := p.getIdentifier()
		if name == nil {
			if !silent {
				p.error(`Expect an identifier, but got %s`, p.peekLexeme())
			}
			return nil
		}

		ids = append(ids, *name)

		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}

	if p.consume(tokenizer.RIGHT_SQ_BRACKET) == nil {
		if !silent {
			p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
		}
		return nil
	}

	return ids
}

func (p *parser) parseParamDecl() ast.Node {
	qualifier := p.consume(tokenizer.CONST, tokenizer.SERIES, tokenizer.SIMPLE)
	begin := p.tell()
	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as type name or argument name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.peekType(0).In(tokenizer.EQUAL, tokenizer.COMMA, tokenizer.RIGHT_PAREN) {
		var def *tokenizer.Token = nil
		if p.consume(tokenizer.EQUAL) != nil {
			def = p.peek(0)
			if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
				p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
				return nil
			}
		}
		beginLoc := name.Begin
		if qualifier != nil {
			beginLoc = qualifier.Begin
		}

		endLoc := name.End
		if def != nil {
			endLoc = def.End
		}
		return ast.WithRange(&ast.ParamDecl{
			Qualifier: pickLexeme(qualifier),
			Type:      nil,
			Name:      name.Lexeme,
			Default:   pickLexeme(def),
		}, beginLoc, endLoc)
	}

	p.seek(begin)
	t := p.parseType(false)
	if t == nil {
		return nil
	}

	name = p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as param name, but got %s`, p.peekLexeme())
		return nil
	}

	var def *tokenizer.Token = nil
	if p.consume(tokenizer.EQUAL) != nil {
		def = p.peek(0)
		if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
			p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
			return nil
		}
		p.consume()
	}

	beginLoc := t.Begin()
	if qualifier != nil {
		beginLoc = qualifier.Begin
	}

	endLoc := name.End
	if def != nil {
		endLoc = def.End
	}
	return ast.WithRange(&ast.ParamDecl{
		Qualifier: pickLexeme(qualifier),
		Type:      t,
		Name:      name.Lexeme,
		Default:   pickLexeme(def),
	}, beginLoc, endLoc)
}

func (p *parser) parseParamList() []*ast.ParamDecl {
	if p.consume(tokenizer.LEFT_PAREN) == nil {
		p.error(`Expect "(", but got %s`, p.peekLexeme())
		return nil
	}

	params := []*ast.ParamDecl{}
	for {
		tt := p.peekType(0)
		if tt == tokenizer.UNKNOWN || tt == tokenizer.RIGHT_PAREN {
			break
		}
		param := p.parseParamDecl()
		if param == nil {
			return nil
		}

		pd, ok := param.(*ast.ParamDecl)
		if !ok {
			p.error(`Unexpected error: param is %T`, param)
			return nil
		}
		params = append(params, pd)

		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}

	if p.consume(tokenizer.RIGHT_PAREN) == nil {
		p.error(`Expect ")" to match "(", but got %s`, p.peekLexeme())
		return nil
	}

	return params
}

func (p *parser) parseFuncDeclStmt() ast.Node {
	export := p.consume(tokenizer.EXPORT)
	method := p.consume(tokenizer.METHOD)

	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as function name, but got %s`, p.peekLexeme())
		return nil
	}

	beginLoc := name.Begin
	if method != nil {
		beginLoc = method.Begin
	}

	if export != nil {
		beginLoc = export.Begin
	}

	params := p.parseParamList()
	if params == nil {
		return nil
	}

	if p.consume(tokenizer.RIGHT_FAT_ARROW) == nil {
		p.error(`Expect "=>", but got %s`, p.peekLexeme())
		return nil
	}

	body := p.parseSuite()
	if body == nil {
		return nil
	}

	return ast.WithRange(&ast.FuncDeclStmt{
		Export: export != nil,
		Method: method != nil,
		Name:   name.Lexeme,
		Params: params,
		Body:   body,
	}, beginLoc, body.End())
}

func (p *parser) parseMemberDecl() ast.Node {
	begin := p.tell()
	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as member name or type name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.peekType(0).In(tokenizer.EQUAL, tokenizer.NEWLINE, tokenizer.DEDENT) {
		var def *tokenizer.Token = nil
		if p.consume(tokenizer.EQUAL) != nil {
			def = p.peek(0)
			if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
				p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
				return nil
			}
			p.consume()
		}
		endLoc := name.End
		if def != nil {
			endLoc = def.End
		}
		return ast.WithRange(&ast.MemberDecl{
			Type:    nil,
			Name:    name.Lexeme,
			Default: pickLexeme(def),
		}, name.Begin, endLoc)
	}

	p.seek(begin)
	t := p.parseType(false)
	if t == nil {
		return nil
	}

	name = p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as member name, but got %s`, p.peekLexeme())
		return nil
	}

	var def *tokenizer.Token = nil
	if p.consume(tokenizer.EQUAL) != nil {
		def = p.peek(0)
		if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
			p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
			return nil
		}
		p.consume()
	}

	endLoc := name.End
	if def != nil {
		endLoc = def.End
	}
	return ast.WithRange(&ast.MemberDecl{
		Type:    t,
		Name:    name.Lexeme,
		Default: pickLexeme(def),
	}, t.Begin(), endLoc)
}

func (p *parser) parseTypeDeclStmt() ast.Node {
	typeToken := p.consume(tokenizer.TYPE)
	if typeToken == nil {
		p.error(`Expect "type", but got %s`, p.peekLexeme())
		return nil
	}

	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifer as user defined type name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.consume(tokenizer.INDENT) == nil {
		p.error(`Expect indent, but got %s`, p.peekLexeme())
		return nil
	}

	endLoc := name.End
	members := []*ast.MemberDecl{}
	for {
		token := p.peek(0)
		if token != nil && token.Type == tokenizer.DEDENT {
			p.consume(tokenizer.DEDENT)
			endLoc = token.End
			break
		}

		member := p.parseMemberDecl()
		if member == nil {
			return nil
		}

		md, ok := member.(*ast.MemberDecl)
		if !ok {
			p.error(`Unexpected error: Stmt to MemberDecl failed`)
			return nil
		}

		members = append(members, md)
		p.consume(tokenizer.NEWLINE)
	}

	return ast.WithRange(&ast.TypeDeclStmt{
		Name:    name.Lexeme,
		Members: members,
	}, typeToken.Begin, endLoc)
}

func (p *parser) parseReassignStmt() ast.Node {
	println("reassign")
	lhs := p.parseAtomExpr(false)
	switch lhs.(type) {
	case *ast.Identifier, *ast.AttrExpr:
	default:
		p.error(`Only identifiers or attributes can be reassigned`)
		return nil
	}

	op := p.consume(tokenizer.COLON_EQUAL, tokenizer.PLUS_EQUAL, tokenizer.MINUS_EQUAL, tokenizer.STAR_EQUAL, tokenizer.SLASH_EQUAL, tokenizer.PERCENT_EQUAL)
	if op == nil {
		p.error(`Expect ":=", "+=", "-=", "*=", "/=" or "%%=", but got %s`, p.peekLexeme())
		return nil
	}

	println("a")
	rhs := p.parseStmt()
	println("b")
	fmt.Println(rhs)

	if rhs == nil {
		return nil
	}

	return ast.WithRange(&ast.ReassignStmt{
		Target: lhs,
		Op:     op.Lexeme,
		Value:  rhs,
	}, lhs.Begin(), rhs.End())
}

func (p *parser) parseImportStmt() ast.Node {
	importToken := p.consume(tokenizer.IMPORT)
	if importToken == nil {
		p.error(`Expect "import", but got %s`, p.peekLexeme())
		return nil
	}

	user := p.getIdentifier()
	if user == nil {
		p.error(`Expect an identifier as author name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.consume(tokenizer.SLASH) == nil {
		p.error(`Expect "/", but got %s`, p.peekLexeme())
		return nil
	}

	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as library name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.consume(tokenizer.SLASH) == nil {
		p.error(`Expect "/", but got %s`, p.peekLexeme())
		return nil
	}

	version := p.consume(tokenizer.IDENTIFIER, tokenizer.NUMBER)
	if version == nil || version.Type != tokenizer.IDENTIFIER && version.Type != tokenizer.NUMBER {
		p.error(`Expect an identifier or a number as library version, but got %s`, p.peekLexeme())
		return nil
	}

	endLoc := version.End

	var alias *tokenizer.Token = nil
	if p.consume(tokenizer.AS) != nil {
		alias = p.getIdentifier()
		if alias == nil {
			p.error(`Expect an identifier as alias of library, but got %s`, p.peekLexeme())
			return nil
		}
		endLoc = alias.End
	}

	return ast.WithRange(&ast.ImportStmt{
		User:    user.Lexeme,
		Name:    name.Lexeme,
		Version: version.Lexeme,
		Alias:   pickLexeme(alias),
	}, importToken.Begin, endLoc)
}

func (p *parser) parseStmt() ast.Node {
	tkn := p.peek(0)
	switch tkn.Type {
	case tokenizer.BREAK:
		p.consume(tokenizer.BREAK)
		return ast.WithRange(&ast.BreakStmt{}, tkn.Begin, tkn.End)
	case tokenizer.CONTINUE:
		p.consume(tokenizer.CONTINUE)
		return ast.WithRange(&ast.ContinueStmt{}, tkn.Begin, tkn.End)
	case tokenizer.IMPORT:
		return p.parseImportStmt()
	case tokenizer.IF:
		return p.parseIfStmt()
	case tokenizer.WHILE:
		return p.parseWhileStmt()
	case tokenizer.FOR:
		return p.parseForStmt()
	case tokenizer.SWITCH:
		return p.parseSwitchStmt()
	case tokenizer.EXPORT:
		if p.peekType(1) == tokenizer.TYPE && p.peekType(3) == tokenizer.IDENTIFIER {
			return p.parseTypeDeclStmt()
		}
		return p.parseFuncDeclStmt()
	case tokenizer.TYPE:
		return p.parseTypeDeclStmt()
	case tokenizer.METHOD:
		return p.parseFuncDeclStmt()
	case tokenizer.VAR, tokenizer.VARIP, tokenizer.CONST, tokenizer.SIMPLE:
		return p.parseVarDeclStmt()
	case tokenizer.IDENTIFIER:
		if p.peekType(1) == tokenizer.LEFT_PAREN && p.hasTokenBeforeNewLine(tokenizer.RIGHT_FAT_ARROW) {
			return p.parseFuncDeclStmt()
		} else if p.peekType(1) == tokenizer.EQUAL {
			return p.parseVarDeclStmt()
		} else if p.peekType(1).In(tokenizer.COLON_EQUAL, tokenizer.PLUS_EQUAL, tokenizer.MINUS_EQUAL, tokenizer.STAR_EQUAL, tokenizer.SLASH_EQUAL, tokenizer.PERCENT_EQUAL) {
			return p.parseReassignStmt()
		} else {
			begin := p.tell()
			lhs := p.parseTestExpr(true)
			afterExpr := p.tell()
			if lhs == nil {
				return nil
			}
			typeSatisfy := false
			switch lhs.(type) {
			case *ast.Identifier, *ast.AttrExpr:
				typeSatisfy = true
			}
			if typeSatisfy {
				token := p.consume(tokenizer.EQUAL, tokenizer.COLON_EQUAL, tokenizer.PLUS_EQUAL, tokenizer.MINUS_EQUAL, tokenizer.STAR_EQUAL, tokenizer.SLASH_EQUAL, tokenizer.PERCENT_EQUAL)
				if token != nil {
					p.seek(begin)
					if token.Type == tokenizer.EQUAL {
						return p.parseVarDeclStmt()
					} else {
						return p.parseReassignStmt()
					}
				}
			}
			p.seek(begin)
			ltype := p.parseType(true)
			if ltype != nil {
				name := p.getIdentifier()
				if name != nil && p.consume(tokenizer.EQUAL) != nil {
					p.seek(begin)
					return p.parseVarDeclStmt()
				}
			}
			p.seek(afterExpr)
			return ast.WithRange(&ast.ExprStmt{
				Expr: lhs,
			}, lhs.Begin(), lhs.End())
		}
	case tokenizer.LEFT_SQ_BRACKET:
		begin := p.tell()
		ids := p.parseIdentifierTuple(true)
		if ids == nil || p.consume(tokenizer.EQUAL) == nil {
			p.seek(begin)
			expr := p.parseTestExpr(false)
			if expr == nil {
				return nil
			}
			return ast.WithRange(&ast.ExprStmt{
				Expr: expr,
			}, expr.Begin(), expr.End())
		}
		init := p.parseStmt()
		if init == nil {
			return nil
		}
		idNames := []string{}
		for _, v := range ids {
			idNames = append(idNames, v.Lexeme)
		}
		return ast.WithRange(&ast.TupleDeclStmt{
			Variables: idNames,
			Initial:   init,
		}, tkn.Begin, init.End())
	default:
		expr := p.parseTestExpr(false)
		if expr == nil {
			return nil
		}
		return ast.WithRange(&ast.ExprStmt{
			Expr: expr,
		}, expr.Begin(), expr.End())
	}
}

func (p *parser) parseStmtGroup() ast.Node {
	stmts := []ast.Node{}
	for {
		stmt := p.parseStmt()
		if stmt == nil {
			return nil
		}
		stmts = append(stmts, stmt)
		if p.consume(tokenizer.COMMA) == nil {
			break
		}
	}
	if len(stmts) == 1 {
		return stmts[0]
	}

	return ast.WithRange(&ast.Suite{
		Body: stmts,
	}, stmts[0].Begin(), stmts[len(stmts)-1].End())
}

func (p *parser) parseSuite() ast.Node {
	indent := p.consume(tokenizer.INDENT)
	if indent == nil {
		// single statement
		stmt := p.parseStmtGroup()
		if stmt == nil {
			return nil
		}
		return stmt
	}

	suite := &ast.Suite{
		Body: []ast.Node{},
	}
	suite.SetBegin(indent.Begin)
	for {
		dedent := p.consume(tokenizer.DEDENT)
		if dedent != nil {
			suite.SetEnd(dedent.End)
			break
		}

		ss := p.parseStmtGroup()
		if ss == nil {
			return nil
		}

		suite.Body = append(suite.Body, ss)
		p.consume(tokenizer.NEWLINE)
	}

	if len(suite.Body) == 1 {
		return suite.Body[0]
	}

	return suite
}

func Parse(tokens []tokenizer.Token) ([]ast.Node, []ParseError) {
	p := parser{
		tokens:  tokens,
		current: 0,
		errors:  []ParseError{},
	}

	stmts := []ast.Node{}
	for !p.eof() {
		stmt := p.parseStmtGroup()
		if stmt == nil {
			p.seekToType(tokenizer.NEWLINE)
			p.consume(tokenizer.NEWLINE)
		} else {
			stmts = append(stmts, stmt)
			p.consume(tokenizer.NEWLINE)
		}
	}

	return stmts, p.errors
}
