package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kvarenzn/pinecone/tokenizer"
)

func parseNumber(token tokenizer.Token) Expr {
	if token.Type != tokenizer.NUMBER {
		return nil
	}
	lexeme := token.Lexeme
	if strings.Contains(lexeme, ".") {
		num, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			return nil
		}
		return FloatLiteral{
			Value: num,
		}
	}
	num, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		return nil
	}
	return IntLiteral{
		Value: num,
	}
}

func parseString(token tokenizer.Token) Expr {
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

	return StringLiteral{
		Value: str,
	}
}

func parseColor(token tokenizer.Token) Expr {
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
	switch len(lexeme) {
	case 4: // #RGB
		return ColorLiteral{
			R: float64((num>>8)&0xf) * 0x11,
			G: float64((num>>4)&0xf) * 0x11,
			B: float64((num>>0)&0xf) * 0x11,
			T: 0.0,
		}
	case 5: // #RGBA
		return ColorLiteral{
			R: float64((num>>12)&0xf) * 0x11,
			G: float64((num>>8)&0xf) * 0x11,
			B: float64((num>>4)&0xf) * 0x11,
			T: 100 - float64((num>>0)&0xf)/0xf*100,
		}
	case 7: // #RRGGBB
		return ColorLiteral{
			R: float64((num >> 16) & 0xff),
			G: float64((num >> 8) & 0xff),
			B: float64((num >> 0) & 0xff),
			T: 0,
		}
	case 9: // #RRGGBBAA
		return ColorLiteral{
			R: float64((num>>24)&0xff) / 0xff,
			G: float64((num>>16)&0xff) / 0xff,
			B: float64((num>>8)&0xff) / 0xff,
			T: 100 - float64((num>>0)&0xff)/0xff*100,
		}
	default:
		return nil
	}
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
		if t == tokenizer.UNKNOWN || t == tokenizer.NEWLINE || t == tokenizer.INDENT || t == tokenizer.DEDENT {
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
			Row: token.Row,
			Col: token.Col,
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

func (p *parser) parseType(silent bool) Type {
	name := p.getIdentifier()
	if name == nil {
		if !silent {
			p.error("Expect a type, but got %s", p.peekLexeme())
		}
		return nil
	}
	var t Type = SimpleType{
		Name: *name,
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
			if p.consume(tokenizer.RIGHT_ANG_BRACKET) == nil {
				if !silent {
					p.error(`Expect ">" to match "<", but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = GenericType{
				Name: t,
				Args: types,
			}
		case tokenizer.DOT:
			name := p.getIdentifier()
			if name == nil {
				if !silent {
					p.error(`Expect an identifier, but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = SubType{
				Name:   t,
				Member: *name,
			}
		case tokenizer.LEFT_SQ_BRACKET:
			// type[] => array<type>
			if p.consume(tokenizer.RIGHT_SQ_BRACKET) == nil {
				if !silent {
					p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
				}
				return nil
			}
			t = GenericType{
				Name: SimpleType{
					Name: tokenizer.Token{
						Type:   tokenizer.IDENTIFIER,
						Lexeme: "array",
					},
				},
				Args: []Type{t},
			}
		}
	}

	return t
}

func (p *parser) parseTypeArgList(silent bool) []Type {
	typeArgs := []Type{}
	for {
		tt := p.peekType(0)
		if tt == tokenizer.UNKNOWN || tt == tokenizer.RIGHT_ANG_BRACKET || tt == tokenizer.NEWLINE {
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

func (p *parser) parseTupleAtom(silent bool) Expr {
	if p.consume(tokenizer.LEFT_SQ_BRACKET) == nil {
		if !silent {
			p.error(`Expect "[", but got %s`, p.peekLexeme())
		}
		return nil
	}
	items := TupleExpr{
		Items: []Expr{},
	}

	for {
		tt := p.peekType(0)
		if tt == tokenizer.UNKNOWN || tt == tokenizer.RIGHT_SQ_BRACKET || tt == tokenizer.NEWLINE {
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

	if p.consume(tokenizer.RIGHT_SQ_BRACKET) == nil {
		if !silent {
			p.error(`Expect "]" to match "[", but got %s`, p.peekLexeme())
		}
		return nil
	}

	return items
}

func (p *parser) parseParenExpr(silent bool) Expr {
	if p.consume(tokenizer.LEFT_PAREN) == nil {
		if !silent {
			p.error(`Expect "(", but got %s`, p.peekLexeme())
		}
		return nil
	}
	expr := p.parseTestExpr(silent)
	if expr == nil {
		return nil
	}
	if p.consume(tokenizer.RIGHT_PAREN) == nil {
		if !silent {
			p.error(`Expect ")" to match "(", but got %s`, p.peekLexeme())
		}
	}

	return expr
}

func (p *parser) parseAtom(silent bool) Expr {
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
		p.consume(tokenizer.COLOR)
		return parseColor(*token)
	case tokenizer.TRUE:
		p.consume(tokenizer.TRUE)
		return TrueExpr{}
	case tokenizer.FALSE:
		p.consume(tokenizer.FALSE)
		return FalseExpr{}
	case tokenizer.LEFT_PAREN:
		return p.parseParenExpr(silent)
	case tokenizer.LEFT_SQ_BRACKET:
		return p.parseTupleAtom(silent)
	default:
		id := p.getIdentifier()
		if id != nil {
			return Identifier{
				Name: *id,
			}
		}
	}

	if !silent {
		p.error(`Expect an identifier, a number, string, color, bool, paren or tuple, but got %s`, p.peekLexeme())
	}

	return nil
}

func (p *parser) parseArgument(silent bool) Expr {
	value := p.parseTestExpr(silent)
	if value == nil {
		return nil
	}
	if p.consume(tokenizer.EQUAL) != nil {
		name := value
		value = p.parseTestExpr(silent)
		if value == nil {
			return nil
		}
		switch n := name.(type) {
		case Identifier:
			return KwArg{
				Name:  n.Name,
				Value: value,
			}
		default:
			return nil
		}
	}
	return value
}

func (p *parser) parseArgList(silent bool) []Expr {
	args := []Expr{}
	for {
		token := p.peek(0)
		if token == nil || token.Type == tokenizer.RIGHT_PAREN || token.Type == tokenizer.NEWLINE {
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

func (p *parser) parseAtomExpr(silent bool) Expr {
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
			if p.consume(tokenizer.RIGHT_PAREN) == nil {
				if !silent {
					p.error(`Expect ")" to match "(", but got %s`, p.peekLexeme())
				}
				return nil
			}
			atom = CallExpr{
				Func: atom,
				Args: args,
			}
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
			p.consume(tokenizer.RIGHT_ANG_BRACKET)
			atom = InstantiationExpr{
				Template: atom,
				TypeArgs: typeArgs,
			}
		} else if p.consume(tokenizer.DOT) != nil {
			member := p.getIdentifier()
			if member == nil {
				if !silent {
					p.error(`Expect an identifier as member name, but got %s`, p.peekLexeme())
				}
				return nil
			}
			atom = AttrExpr{
				Target: atom,
				Name:   *member,
			}
		} else {
			break
		}
	}

	return atom
}

func (p *parser) parseArithmeticFactor(silent bool) Expr {
	op := p.consume(tokenizer.PLUS, tokenizer.MINUS)
	if op != nil {
		expr := p.parseArithmeticFactor(silent)
		if expr == nil {
			return nil
		}
		return UnaryExpr{
			Op:   *op,
			Expr: expr,
		}
	}
	expr := p.parseAtomExpr(silent)
	if expr == nil {
		return nil
	}
	return expr
}

func (p *parser) parseArithmeticTerm(silent bool) Expr {
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
		left = BinaryExpr{
			Left:  left,
			Op:    *op,
			Right: right,
		}
	}
	return left
}

func (p *parser) parseArithmeticExpr(silent bool) Expr {
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
		left = BinaryExpr{
			Left:  left,
			Op:    *op,
			Right: right,
		}
	}
	return left
}

func (p *parser) parseComparison(silent bool) Expr {
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
		left = BinaryExpr{
			Left:  left,
			Op:    *op,
			Right: right,
		}
	}
	return left
}

func (p *parser) parseNotTest(silent bool) Expr {
	not := p.consume(tokenizer.NOT)
	if not == nil {
		return p.parseComparison(silent)
	}
	notExpr := p.parseNotTest(silent)
	if notExpr == nil {
		return nil
	}
	return UnaryExpr{
		Op:   *not,
		Expr: notExpr,
	}
}

func (p *parser) parseAndTest(silent bool) Expr {
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
		left = BinaryExpr{
			Left:  left,
			Op:    *and,
			Right: right,
		}
	}
	return left
}

func (p *parser) parseOrTest(silent bool) Expr {
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
		left = BinaryExpr{
			Left:  left,
			Op:    *or,
			Right: right,
		}
	}
	return left
}

func (p *parser) parseTestExpr(silent bool) Expr {
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
	return TernaryExpr{
		Test:  test,
		True:  t,
		False: f,
	}
}

func (p *parser) parseIfStmt() Stmt {
	if p.consume(tokenizer.IF) == nil {
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

	ifStmt := IfStmt{
		Test: test,
		True: suite,
	}

	if p.consume(tokenizer.ELSE) != nil {
		f := p.parseSuite()
		if f == nil {
			return nil
		}
		ifStmt.False = f
	}

	return ifStmt
}

func (p *parser) parseWhileStmt() Stmt {
	if p.consume(tokenizer.WHILE) == nil {
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
	return WhileStmt{
		Test: test,
		Body: suite,
	}
}

func (p *parser) parseForStmt() Stmt {
	if p.consume(tokenizer.FOR) == nil {
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

		return ForInStmt{
			Index:     *idx,
			Iterator:  *iter,
			Container: container,
			Body:      suite,
		}
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

		var step Expr = nil
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

		return ForStmt{
			Counter: *counter,
			Init:    init,
			Step:    step,
			Final:   final,
			Body:    suite,
		}
	} else if p.consume(tokenizer.IN) != nil {
		container := p.parseTestExpr(false)
		if container == nil {
			return nil
		}

		suite := p.parseSuite()
		if suite == nil {
			return nil
		}

		return ForInStmt{
			Iterator:  *counter,
			Container: container,
			Body:      suite,
		}
	}

	p.error(`Expect "in" or "=", but got %s`, p.peekLexeme())
	return nil
}

func (p *parser) parseCaseClause() Stmt {
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

	return CaseClause{
		Cond: cond,
		Body: body,
	}
}

func (p *parser) parseSwitchStmt() Stmt {
	if p.consume(tokenizer.SWITCH) == nil {
		p.error(`Expect "switch", but got %s`, p.peekLexeme())
		return nil
	}

	var target Expr = nil

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

	switchStmt := SwitchStmt{
		Target:  target,
		Cases:   []CaseClause{},
		Default: nil,
	}

	for {
		token := p.peek(0)
		if token == nil {
			p.error(`Uncomplete switch statement`)
			return nil
		}

		if token.Type == tokenizer.DEDENT {
			p.consume(tokenizer.DEDENT)
			break
		}

		caseClause := p.parseCaseClause()
		if caseClause == nil {
			return nil
		}

		switch c := caseClause.(type) {
		case CaseClause:
			switchStmt.Cases = append(switchStmt.Cases, c)
		default:
			switchStmt.Default = c
		}
		p.consume(tokenizer.NEWLINE)
	}
	return switchStmt
}

func (p *parser) parseVarDeclStmt() Stmt {
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
		return VarDeclStmt{
			DeclMode:  declMode,
			Qualifier: qualifier,
			Type:      nil,
			Name:      *name,
			Initial:   init,
		}
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

	return VarDeclStmt{
		DeclMode:  declMode,
		Qualifier: qualifier,
		Type:      t,
		Name:      *name,
		Initial:   init,
	}
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

func (p *parser) parseParamDecl() Stmt {
	qualifier := p.consume(tokenizer.CONST, tokenizer.SERIES, tokenizer.SIMPLE)
	begin := p.tell()
	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as type name or argument name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.peekType(0) == tokenizer.EQUAL || p.peekType(0) == tokenizer.COMMA || p.peekType(0) == tokenizer.RIGHT_PAREN {
		var def *tokenizer.Token = nil
		if p.consume(tokenizer.EQUAL) != nil {
			def = p.peek(0)
			if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
				p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
				return nil
			}
		}
		return ParamDecl{
			Qualifier: qualifier,
			Type:      nil,
			Name:      *name,
			Default:   def,
		}
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

	return ParamDecl{
		Qualifier: qualifier,
		Type:      t,
		Name:      *name,
		Default:   def,
	}
}

func (p *parser) parseParamList() []ParamDecl {
	if p.consume(tokenizer.LEFT_PAREN) == nil {
		p.error(`Expect "(", but got %s`, p.peekLexeme())
		return nil
	}

	params := []ParamDecl{}
	for {
		tt := p.peekType(0)
		if tt == tokenizer.UNKNOWN || tt == tokenizer.RIGHT_PAREN {
			break
		}
		param := p.parseParamDecl()
		if param == nil {
			return nil
		}

		switch pd := param.(type) {
		case ParamDecl:
			params = append(params, pd)
		default:
			p.error(`Unexpected error`)
			return nil
		}

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

func (p *parser) parseFuncDeclStmt() Stmt {
	export := false
	method := false
	if p.consume(tokenizer.EXPORT) != nil {
		export = true
	}

	if p.consume(tokenizer.METHOD) != nil {
		method = true
	}

	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as function name, but got %s`, p.peekLexeme())
		return nil
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

	return FuncDeclStmt{
		Export: export,
		Method: method,
		Name:   *name,
		Params: params,
		Body:   body,
	}
}

func (p *parser) parseMemberDecl() Stmt {
	begin := p.tell()
	name := p.getIdentifier()
	if name == nil {
		p.error(`Expect an identifier as member name or type name, but got %s`, p.peekLexeme())
		return nil
	}

	if p.peekType(0) == tokenizer.EQUAL || p.peekType(0) == tokenizer.NEWLINE || p.peekType(0) == tokenizer.DEDENT {
		var def *tokenizer.Token = nil
		if p.consume(tokenizer.EQUAL) != nil {
			def = p.peek(0)
			if def == nil || !(def.Type == tokenizer.IDENTIFIER || def.IsSoftKeyword() || def.IsLiteral()) {
				p.error(`Expect an identifier or a literal, but got %s`, p.peekLexeme())
				return nil
			}
			p.consume()
		}
		return MemberDecl{
			Type:    nil,
			Name:    *name,
			Default: def,
		}
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

	return MemberDecl{
		Type:    t,
		Name:    *name,
		Default: def,
	}
}

func (p *parser) parseTypeDeclStmt() Stmt {
	if p.consume(tokenizer.TYPE) == nil {
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

	members := []MemberDecl{}
	for {
		token := p.peek(0)
		if token != nil && token.Type == tokenizer.DEDENT {
			p.consume(tokenizer.DEDENT)
			break
		}

		member := p.parseMemberDecl()
		if member == nil {
			return nil
		}

		md, ok := member.(MemberDecl)
		if !ok {
			p.error(`Unexpected error: Stmt to MemberDecl failed`)
			return nil
		}

		members = append(members, md)
		p.consume(tokenizer.NEWLINE)
	}

	return TypeDeclStmt{
		Name:    *name,
		Members: members,
	}
}

func (p *parser) parseReassignStmt() Stmt {
	println("reassign")
	lhs := p.parseAtomExpr(false)
	switch lhs.(type) {
	case Identifier:
	case AttrExpr:
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

	return ReassignStmt{
		Target: lhs,
		Op:     *op,
		Value:  rhs,
	}
}

func (p *parser) parseImportStmt() Stmt {
	if p.consume(tokenizer.IMPORT) == nil {
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

	var alias *tokenizer.Token = nil
	if p.consume(tokenizer.AS) != nil {
		alias = p.getIdentifier()
		if alias == nil {
			p.error(`Expect an identifier as alias of library, but got %s`, p.peekLexeme())
			return nil
		}
	}

	return ImportStmt{
		User:    *user,
		Name:    *name,
		Version: *version,
		Alias:   alias,
	}
}

func (p *parser) parseStmt() Stmt {
	switch p.peekType(0) {
	case tokenizer.BREAK:
		p.consume(tokenizer.BREAK)
		return BreakStmt{}
	case tokenizer.CONTINUE:
		p.consume(tokenizer.CONTINUE)
		return ContinueStmt{}
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
	case tokenizer.VAR:
		fallthrough
	case tokenizer.VARIP:
		fallthrough
	case tokenizer.CONST:
		fallthrough
	case tokenizer.SERIES:
		fallthrough
	case tokenizer.SIMPLE:
		return p.parseVarDeclStmt()
	case tokenizer.IDENTIFIER:
		if p.peekType(1) == tokenizer.LEFT_PAREN && p.hasTokenBeforeNewLine(tokenizer.RIGHT_FAT_ARROW) {
			return p.parseFuncDeclStmt()
		} else if p.peekType(1) == tokenizer.EQUAL {
			return p.parseVarDeclStmt()
		} else if p.peekType(1) == tokenizer.COLON_EQUAL || p.peekType(1) == tokenizer.PLUS_EQUAL || p.peekType(1) == tokenizer.MINUS_EQUAL || p.peekType(1) == tokenizer.STAR_EQUAL || p.peekType(1) == tokenizer.SLASH_EQUAL || p.peekType(1) == tokenizer.PERCENT_EQUAL {
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
			case Identifier:
				typeSatisfy = true
			case AttrExpr:
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
			return ExprStmt{
				Expr: lhs,
			}
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
			return ExprStmt{
				Expr: expr,
			}
		}
		init := p.parseStmt()
		if init == nil {
			return nil
		}
		return TupleDeclStmt{
			Variables: ids,
			Initial:   init,
		}
	default:
		expr := p.parseTestExpr(false)
		if expr == nil {
			return nil
		}
		return ExprStmt{
			Expr: expr,
		}
	}
}

func (p *parser) parseStmtGroup() Stmt {
	stmts := []Stmt{}
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

	return Suite{
		Body: stmts,
	}
}

func (p *parser) parseSuite() Stmt {
	if p.consume(tokenizer.INDENT) == nil {
		// single statement
		stmt := p.parseStmtGroup()
		if stmt == nil {
			return nil
		}
		return stmt
	}

	stmts := []Stmt{}
	for {
		token := p.peek(0)
		if token != nil && token.Type == tokenizer.DEDENT {
			p.consume(tokenizer.DEDENT)
			break
		}

		ss := p.parseStmtGroup()
		if ss == nil {
			return nil
		}

		stmts = append(stmts, ss)
		p.consume(tokenizer.NEWLINE)
	}

	if len(stmts) == 1 {
		return stmts[0]
	}

	return Suite{
		Body: stmts,
	}
}

func Parse(tokens []tokenizer.Token) ([]Stmt, []ParseError) {
	p := parser{
		tokens:  tokens,
		current: 0,
		errors:  []ParseError{},
	}

	stmts := []Stmt{}
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
