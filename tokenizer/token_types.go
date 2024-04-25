package tokenizer

type TokenType byte

const (
	UNKNOWN TokenType = iota

	metaBegin
	NEWLINE
	INDENT
	DEDENT
	metaEnd

	delimiterBegin
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_SQ_BRACKET
	RIGHT_SQ_BRACKET
	LEFT_ANG_BRACKET
	RIGHT_ANG_BRACKET
	COMMA
	delimiterEnd

	operatorBegin
	DOT
	QUESTION
	COLON
	EQUAL
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	COLON_EQUAL
	EQUAL_EQUAL
	BANG_EQUAL
	GREATER_EQUAL
	LESS_EQUAL
	PLUS_EQUAL
	MINUS_EQUAL
	STAR_EQUAL
	SLASH_EQUAL
	PERCENT_EQUAL
	RIGHT_FAT_ARROW
	operatorEnd

	IDENTIFIER

	literalBegin
	STRING
	NUMBER
	COLOR
	literalEnd

	keywordBegin
	EXPORT
	IMPORT
	AS
	TYPE
	METHOD
	AND
	OR
	NOT
	IF
	ELSE
	FOR
	TO
	BY
	IN
	WHILE
	SWITCH
	BREAK
	CONTINUE

	CATCH
	CLASS
	DO
	ELLIPSE
	IS
	POLYGON
	RANGE
	RETURN
	STRUCT
	TEXT
	THROW
	TRY

	// type qualifiers
	CONST
	SIMPLE
	SERIES

	// declaration modes
	VARIP
	VAR

	TRUE
	FALSE

	keywordEnd
)

var TOKEN_TYPE_NAMES = map[TokenType]string{
	UNKNOWN:           "UNKNOWN",
	metaBegin:         "metaBegin",
	NEWLINE:           "NEWLINE",
	INDENT:            "INDENT",
	DEDENT:            "DEDENT",
	metaEnd:           "metaEnd",
	delimiterBegin:    "delimiterBegin",
	LEFT_PAREN:        "LEFT_PAREN",
	RIGHT_PAREN:       "RIGHT_PAREN",
	LEFT_SQ_BRACKET:   "LEFT_SQ_BRACKET",
	RIGHT_SQ_BRACKET:  "RIGHT_SQ_BRACKET",
	LEFT_ANG_BRACKET:  "LEFT_ANG_BRACKET",
	RIGHT_ANG_BRACKET: "RIGHT_ANG_BRACKET",
	COMMA:             "COMMA",
	delimiterEnd:      "delimiterEnd",
	operatorBegin:     "operatorBegin",
	DOT:               "DOT",
	QUESTION:          "QUESTION",
	COLON:             "COLON",
	EQUAL:             "EQUAL",
	PLUS:              "PLUS",
	MINUS:             "MINUS",
	STAR:              "STAR",
	SLASH:             "SLASH",
	PERCENT:           "PERCENT",
	COLON_EQUAL:       "COLON_EQUAL",
	EQUAL_EQUAL:       "EQUAL_EQUAL",
	BANG_EQUAL:        "BANG_EQUAL",
	GREATER_EQUAL:     "GREATER_EQUAL",
	LESS_EQUAL:        "LESS_EQUAL",
	PLUS_EQUAL:        "PLUS_EQUAL",
	MINUS_EQUAL:       "MINUS_EQUAL",
	STAR_EQUAL:        "STAR_EQUAL",
	SLASH_EQUAL:       "SLASH_EQUAL",
	PERCENT_EQUAL:     "PERCENT_EQUAL",
	RIGHT_FAT_ARROW:   "RIGHT_FAT_ARROW",
	operatorEnd:       "operatorEnd",
	IDENTIFIER:        "IDENTIFIER",
	literalBegin:      "literalBegin",
	STRING:            "STRING",
	NUMBER:            "NUMBER",
	COLOR:             "COLOR",
	literalEnd:        "literalEnd",
	keywordBegin:      "keywordBegin",
	EXPORT:            "EXPORT",
	IMPORT:            "IMPORT",
	AS:                "AS",
	TYPE:              "TYPE",
	METHOD:            "METHOD",
	AND:               "AND",
	OR:                "OR",
	NOT:               "NOT",
	IF:                "IF",
	ELSE:              "ELSE",
	FOR:               "FOR",
	TO:                "TO",
	BY:                "BY",
	IN:                "IN",
	WHILE:             "WHILE",
	SWITCH:            "SWITCH",
	BREAK:             "BREAK",
	CONTINUE:          "CONTINUE",
	CATCH:             "CATCH",
	CLASS:             "CLASS",
	DO:                "DO",
	ELLIPSE:           "ELLIPSE",
	IS:                "IS",
	POLYGON:           "POLYGON",
	RANGE:             "RANGE",
	RETURN:            "RETURN",
	STRUCT:            "STRUCT",
	TEXT:              "TEXT",
	THROW:             "THROW",
	TRY:               "TRY",
	CONST:             "CONST",
	SIMPLE:            "SIMPLE",
	SERIES:            "SERIES",
	VARIP:             "VARIP",
	VAR:               "VAR",
	TRUE:              "TRUE",
	FALSE:             "FALSE",
	keywordEnd:        "keywordEnd",
}

var KEYWORDS = map[string]TokenType{
	"export":   EXPORT,
	"import":   IMPORT,
	"as":       AS,
	"type":     TYPE,
	"method":   METHOD,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"to":       TO,
	"by":       BY,
	"in":       IN,
	"while":    WHILE,
	"switch":   SWITCH,
	"break":    BREAK,
	"continue": CONTINUE,

	"catch":   CATCH,
	"class":   CLASS,
	"do":      DO,
	"ellipse": ELLIPSE,
	"is":      IS,
	"polygon": POLYGON,
	"range":   RANGE,
	"return":  RETURN,
	"struct":  STRUCT,
	"text":    TEXT,
	"throw":   THROW,
	"try":     TRY,

	"const":  CONST,
	"simple": SIMPLE,
	"series": SERIES,

	"varip": VARIP,
	"var":   VAR,

	"true":  TRUE,
	"false": FALSE,
}

func (tt TokenType) In(tts ...TokenType) bool {
	for _, t := range tts {
		if t == tt {
			return true
		}
	}
	return false
}
