package ecmascript

const (
	EOFToken = iota
	PunctuatorToken
	StringLiteralToken
	IdentifierToken
	KeywordToken
	NullLiteralToken
	BooleanLiteralToken
	NumericLiteralToken
)

type Token struct {
	token_type  int
	token_value []byte
}

var NULLToken Token = Token{EOFToken, []byte("")}
