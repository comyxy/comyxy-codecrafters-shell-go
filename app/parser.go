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

func (p *Parser) ParsePipeline(sh *Shell) []*Command {
	var cmds []*Command

	for {
		cmd := p.Parse(sh)

		if len(cmd.Args) > 0 {
			cmds = append(cmds, cmd)
		}

		if p.cur.Type == TokenPipeline {
			p.advance()
			continue
		}

		break
	}

	return cmds
}

func (p *Parser) Parse(sh *Shell) *Command {
	cmd := NewCommand(sh)

	for p.cur.Type != TokenEOF && p.cur.Type != TokenPipeline {

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
