package main

import "strings"

type ParseState int

const (
	StateNormal ParseState = iota
	StateSingleQuote
	StateDoubleQuote
)

type Parser struct {
	state ParseState
	res   []string
	sb    strings.Builder
}

func NewParser() *Parser {
	return &Parser{
		state: StateNormal,
		res:   nil,
		sb:    strings.Builder{},
	}
}

func (p *Parser) Parse(input string) []string {
	for _, c := range input {
		p.processRune(c)
	}

	if p.sb.Len() > 0 {
		p.finalizeArg()
	}

	return p.res
}

func (p *Parser) processRune(c rune) {
	switch p.state {
	case StateNormal:
		p.processStateNormal(c)
	case StateSingleQuote:
		p.processStateSingleQuote(c)
	case StateDoubleQuote:
		p.processStateDoubleQuote(c)
	}
}

func (p *Parser) processStateNormal(c rune) {
	switch c {
	case '\'':
		// 单引号
		p.state = StateSingleQuote
	case '"':
		// 双引号
		p.state = StateDoubleQuote
	case ' ':
		p.finalizeArg()
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) processStateSingleQuote(c rune) {
	switch c {
	case '\'':
		p.state = StateNormal
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) processStateDoubleQuote(c rune) {
	switch c {
	case '"':
		p.state = StateNormal
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) finalizeArg() {
	if p.sb.Len() > 0 {
		p.res = append(p.res, p.sb.String())
		p.sb.Reset()
	}
}
