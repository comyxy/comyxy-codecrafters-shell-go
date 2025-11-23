package main

type TokenType int

const (
	TokenEOF TokenType = iota + 1
	TokenWord
	TokenRedirectOut       // 1>
	TokenRedirectOutAppend // 1>>
	TokenRedirectErr       // 2>
	TokenRedirectErrAppend // 2>>
	TokenPipeline          // |
	TokenRedirectIn        // <
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
	case TokenRedirectOutAppend:
		return "REDIRECT_OUT_APPEND"
	case TokenRedirectErr:
		return "REDIRECT_ERR"
	case TokenRedirectErrAppend:
		return "REDIRECT_ERRAPPEND"
	case TokenPipeline:
		return "PIPELINE"
	default:
		return "UNKNOWN"
	}
}
