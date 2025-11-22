package main

type Parser struct {
	tokens []Token
	pos    int
	cur    Token
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.cur = tokens[0]
	}
	return p
}

func (p *Parser) Parse() *Command {
	cmd := &Command{
		Args: []string{},
	}

	for p.cur.Type != TokenEOF {

		curType := p.cur.Type

		switch p.cur.Type {
		case TokenWord:
			cmd.Args = append(cmd.Args, p.cur.Val)
			p.advance()
		case TokenRedirectOut, TokenRedirectOutAppend:
			p.advance()
			if p.cur.Type == TokenWord {
				cmd.RedirectOutput = Redirect{
					TokenType: curType,
					FileName:  p.cur.Val,
				}
				p.advance()
			}
		case TokenRedirectErr, TokenRedirectErrAppend:
			p.advance()
			if p.cur.Type == TokenWord {
				cmd.RedirectErr = Redirect{
					TokenType: curType,
					FileName:  p.cur.Val,
				}
				p.advance()
			}
		default:
			p.advance()
		}
	}

	return cmd
}

func (p *Parser) advance() {
	p.pos += 1
	if p.pos >= len(p.tokens) {
		p.cur = NewToken(TokenEOF, "")
	} else {
		p.cur = p.tokens[p.pos]
	}
}

func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return NewToken(TokenEOF, "")
	}
	return p.tokens[p.pos]
}
