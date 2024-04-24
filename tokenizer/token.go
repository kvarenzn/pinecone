package tokenizer

import (
	"fmt"

	"github.com/kvarenzn/pinecone/structs"
)

type Token struct {
	Type   TokenType
	Lexeme string
	Begin  structs.Location
	End    structs.Location
}

func (t Token) String() string {
	return fmt.Sprintf("<%v: %q>", TOKEN_TYPE_NAMES[t.Type], t.Lexeme)
}

func (t Token) IsSoftKeyword() bool {
	return t.Type == TYPE || t.Type == CATCH || t.Type == CLASS || t.Type == DO || t.Type == ELLIPSE || t.Type == IS || t.Type == POLYGON || t.Type == RANGE || t.Type == RETURN || t.Type == STRUCT || t.Type == TEXT || t.Type == THROW || t.Type == TRY
}

func (t Token) IsLiteral() bool {
	return t.Type == NUMBER || t.Type == STRING || t.Type == COLOR || t.Type == TRUE || t.Type == FALSE
}
