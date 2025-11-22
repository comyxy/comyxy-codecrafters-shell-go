package main

type TokenType int

const (
	TokenEOF TokenType = iota + 1
	TokenWord
	TokenRedirectOut       // 1>
	TokenRedirectOutAppend // 1>>
	TokenRedirectErr       // 2>
	TokenRedirectErrAppend // 2>>
)

type Token struct {
	Type TokenType
	Val  string
}

func NewToken(tokenType TokenType, val string) Token {
	return Token{
		Type: tokenType,
		Val:  val,
	}
}

func (t TokenType) String() string {
	switch t {
	case TokenWord:
		return "WORD"
	case TokenRedirectOut:
		return "REDIRECT_OUT"
	default:
		return "UNKNOWN"
	}
}
