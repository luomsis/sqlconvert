package sqlconvert

// Token types
const (
	TokenWord      = iota + 1 // Keyword or unquoted identifier (will be refined later)
	TokenKeyword              // Keyword (part of language, not only reserved words)
	TokenIdent                // Identifier (quoted, unquoted, qualified multi-part)
	TokenSymbol               // Single char symbol
	TokenNumber               // Number
	TokenString               // Quoted string
	TokenComment              // Comment
	TokenBOM                  // UTF byte order mark
	TokenFunction             // Function
	TokenStatement            // Statement
)

type Token struct {
	tokenType int
	open      *Token
	close     *Token
	prev      *Token
	next      *Token
	content   []byte
}

func newToken() *Token {
	return &Token{
		content: make([]byte, 0),
	}
}

func (t *Token) IsBlank() bool {
	return false
}

func IsNumberic() bool {
	return false
}
