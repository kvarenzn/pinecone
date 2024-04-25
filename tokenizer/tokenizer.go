package tokenizer

import (
	"fmt"

	"github.com/kvarenzn/pinecone/structs"
)

func isDecimal(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHex(r rune) bool {
	return isDecimal(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isIdentifierFirstRune(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isIdentifierRune(r rune) bool {
	return isIdentifierFirstRune(r) || isDecimal(r)
}

type tokenizer struct {
	source     []rune
	start      int
	current    int
	startRow   int
	startCol   int
	currentRow int
	currentCol int
	prevRow    int
	prevCol    int
	tokens     []Token
	indents    []int
}

const eof rune = -1

func (ts tokenizer) eof() bool {
	return ts.current >= len(ts.source)
}

func (ts tokenizer) peek(n int) rune {
	if ts.current+n >= len(ts.source) {
		return eof
	}

	return ts.source[ts.current+n]
}

func (ts *tokenizer) consume() {
	if ts.eof() {
		return
	}

	ts.prevCol = ts.currentCol
	ts.prevRow = ts.currentRow

	if ts.peek(0) == '\n' || ts.peek(0) == '\r' && ts.peek(1) != '\n' {
		ts.currentCol = 1
		ts.currentRow++
	} else {
		ts.currentCol++
	}

	ts.current++
}

func (ts *tokenizer) advance() rune {
	pos := ts.current
	ts.consume()
	return ts.source[pos]
}

func (ts *tokenizer) match(target rune) bool {
	if ts.eof() {
		return false
	}

	if ts.source[ts.current] != target {
		return false
	}

	ts.consume()
	return true
}

func (ts tokenizer) take() string {
	return string(ts.source[ts.start:ts.current])
}

func (ts tokenizer) takeAs(tt TokenType) Token {
	return Token{
		Type:   tt,
		Lexeme: ts.take(),
		Begin: structs.Location{
			Row:    ts.startRow,
			Column: ts.startCol,
		},
		End: structs.Location{
			Row:    ts.prevRow,
			Column: ts.prevCol,
		},
	}
}

func (ts *tokenizer) fastForward() {
	ts.start = ts.current
	ts.startRow = ts.currentRow
	ts.startCol = ts.currentCol
}

func (t *tokenizer) record(tt TokenType) {
	t.tokens = append(t.tokens, t.takeAs(tt))
}

func (t *tokenizer) setCurrentIndent(indent int) {
	if indent%4 != 0 {
		if t.tokens[len(t.tokens)-1].Type == NEWLINE {
			t.tokens = t.tokens[:len(t.tokens)-1]
		}
		return
	}

	top := t.indents[len(t.indents)-1]
	if indent > top {
		t.indents = append(t.indents, indent)
		t.record(INDENT)
		return
	} else if indent < top {
		for i, v := range t.indents {
			if v == indent {
				for j := i + 1; j < len(t.indents); j++ {
					t.record(DEDENT)
				}
				t.indents = t.indents[:i+1]
				return
			}
		}
		panic("Invalid Indent")
	}

	if len(t.tokens) > 0 {
		t.record(NEWLINE)
	}
}

func (t *tokenizer) scanIndent() {
	indent := 0
	for !t.eof() {
		switch t.peek(0) {
		case ' ':
			indent++
			t.advance()
		case '\t':
			indent += 4
			t.advance()
		case '\f':
			t.advance()
		case '/':
			if t.peek(1) == '/' {
				for t.peek(0) != '\r' && t.peek(0) != '\n' && !t.eof() {
					t.advance()
				}
				return
			}
			t.setCurrentIndent(indent)
			return
		case '\r', '\n':
			return
		default:
			t.setCurrentIndent(indent)
			return
		}
	}
}

func (t *tokenizer) atStart() rune {
	return t.source[t.start]
}

func (t *tokenizer) scanString() {
	start := t.atStart()
	for !t.eof() {
		r := t.advance()
		if r == start {
			t.record(STRING)
			break
		} else if r == '\\' {
			t.advance()
		}
	}
}

func (t *tokenizer) scanColor() {
	for !t.eof() {
		r := t.peek(0)
		if !isHex(r) {
			t.record(COLOR)
			return
		}
		t.advance()
	}
}

func (t *tokenizer) scanNumber() {
	fracPart := t.atStart() == '.'
	expoPart := false

	for !t.eof() {
		r := t.peek(0)
		if isDecimal(r) {
			t.advance()
			continue
		}

		if fracPart {
			if !expoPart && (r == 'e' || r == 'E') {
				expoPart = true
			} else {
				t.record(NUMBER)
				return
			}
		} else if r == '.' {
			t.advance()
			fracPart = true
		} else {
			t.record(NUMBER)
			return
		}
	}

	t.record(NUMBER)
}

func (t *tokenizer) scanIdentifier() {
	for isIdentifierRune(t.peek(0)) {
		t.advance()
	}
	lexeme := t.take()
	tt, ok := KEYWORDS[lexeme]
	if ok {
		t.record(tt)
	} else {
		t.record(IDENTIFIER)
	}
}

func (t *tokenizer) scanToken() {
	r := t.advance()
	switch r {
	case '(':
		t.record(LEFT_PAREN)
	case ')':
		t.record(RIGHT_PAREN)
	case '[':
		t.record(LEFT_SQ_BRACKET)
	case ']':
		t.record(RIGHT_SQ_BRACKET)
	case '<':
		if t.match('=') {
			t.record(LESS_EQUAL)
		} else {
			t.record(LEFT_ANG_BRACKET)
		}
	case '>':
		if t.match('=') {
			t.record(GREATER_EQUAL)
		} else {
			t.record(RIGHT_ANG_BRACKET)
		}
	case ',':
		t.record(COMMA)
	case '?':
		t.record(QUESTION)
	case ':':
		if t.match('=') {
			t.record(COLON_EQUAL)
		} else {
			t.record(COLON)
		}
	case '!':
		if t.match('=') {
			t.record(BANG_EQUAL)
		} else {
			t.record(NOT)
		}
	case '=':
		if t.match('>') {
			t.record(RIGHT_FAT_ARROW)
		} else if t.match('=') {
			t.record(EQUAL_EQUAL)
		} else {
			t.record(EQUAL)
		}
	case '+':
		if t.match('=') {
			t.record(PLUS_EQUAL)
		} else {
			t.record(PLUS)
		}
	case '-':
		if t.match('=') {
			t.record(MINUS_EQUAL)
		} else {
			t.record(MINUS)
		}
	case '*':
		if t.match('=') {
			t.record(STAR_EQUAL)
		} else {
			t.record(STAR)
		}
	case '/':
		if t.match('/') {
			for t.peek(0) != '\n' && t.peek(0) != '\r' && !t.eof() {
				t.advance()
			}
		} else if t.match('=') {
			t.record(SLASH_EQUAL)
		} else {
			t.record(SLASH)
		}
	case '%':
		if t.match('=') {
			t.record(PERCENT_EQUAL)
		} else {
			t.record(PERCENT)
		}
	case '\'', '"':
		t.scanString()
	case ' ', '\t':
		return
	case '\r', '\n':
		t.scanIndent()
	case '#':
		t.scanColor()
	default:
		if isDecimal(r) || r == '.' && isDecimal(t.peek(0)) {
			t.scanNumber()
			return
		}
		if r == '.' {
			t.record(DOT)
			return
		}
		if isIdentifierFirstRune(r) {
			t.scanIdentifier()
			return
		}
		panic(fmt.Sprintf("Unexpected character: %#U", r))
	}
}

func Tokenize(source string) []Token {
	t := &tokenizer{
		source:     []rune(source),
		start:      0,
		current:    0,
		startRow:   1,
		startCol:   1,
		prevRow:    1,
		prevCol:    1,
		currentRow: 1,
		currentCol: 1,
		tokens:     []Token{},
		indents:    []int{0},
	}

	if !t.eof() {
		t.fastForward()
		t.scanIndent()
	}

	for !t.eof() {
		t.fastForward()
		t.scanToken()
	}

	t.fastForward()
	t.setCurrentIndent(0)

	return t.tokens
}
