package main

import "strings"

type ParseState int

const (
	StateNormal ParseState = iota
	StateSingleQuote
	StateDoubleQuote
	StateEscaped
)

type ParseStateStack struct {
	stack []ParseState
}

func NewParseStateStack() *ParseStateStack {
	return &ParseStateStack{stack: []ParseState{StateNormal}}
}

func (s *ParseStateStack) Push(state ParseState) {
	s.stack = append(s.stack, state)
}

func (s *ParseStateStack) Pop() ParseState {
	if len(s.stack) <= 1 {
		// 保留初始状态normal
		return StateNormal
	}
	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return top
}

func (s *ParseStateStack) Peek() ParseState {
	return s.stack[len(s.stack)-1]
}

type Parser struct {
	stateStack *ParseStateStack
	res        []string
	sb         strings.Builder
}

func NewParser() *Parser {
	return &Parser{
		stateStack: NewParseStateStack(),
		res:        nil,
		sb:         strings.Builder{},
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
	currentState := p.stateStack.Peek()

	switch currentState {
	case StateNormal:
		p.processStateNormal(c)
	case StateSingleQuote:
		p.processStateSingleQuote(c)
	case StateDoubleQuote:
		p.processStateDoubleQuote(c)
	case StateEscaped:
		p.processStateEscaped(c)
	}
}

func (p *Parser) processStateNormal(c rune) {
	switch c {
	case '\'':
		// 单引号
		p.stateStack.Push(StateSingleQuote)
	case '"':
		// 双引号
		p.stateStack.Push(StateDoubleQuote)
	case '\\':
		p.stateStack.Push(StateEscaped)
	case ' ':
		p.finalizeArg()
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) processStateSingleQuote(c rune) {
	switch c {
	case '\'':
		p.stateStack.Pop()
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) processStateDoubleQuote(c rune) {
	switch c {
	case '"':
		p.stateStack.Pop()
	case '\\':
		// 双引号支持部分字符转义
		p.stateStack.Push(StateEscaped)
	default:
		p.sb.WriteRune(c)
	}
}

func (p *Parser) processStateEscaped(c rune) {
	p.stateStack.Pop()

	prevState := p.stateStack.Peek()
	switch prevState {
	case StateDoubleQuote:
		if c == '"' || c == '\\' {
			p.sb.WriteRune(c)
		} else {
			p.sb.WriteRune('\\')
			p.sb.WriteRune(c)
		}
	case StateNormal:
		p.sb.WriteRune(c)
	default:
		panic("unreachable")
	}
}

func (p *Parser) finalizeArg() {
	if p.sb.Len() > 0 {
		p.res = append(p.res, p.sb.String())
		p.sb.Reset()
	}
}
