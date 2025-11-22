package main

import "strings"

type Scanner struct {
	input string
	pos   int
	cur   rune
}

func NewScanner(input string) *Scanner {
	scanner := &Scanner{
		input: input,
		pos:   0,
		cur:   0, // rune=0 null unicode
	}
	if len(input) > 0 {
		scanner.cur = rune(input[0])
	}
	return scanner
}

func (sc *Scanner) Scan() []string {
	var res []string

	for sc.cur != 0 {

		switch sc.cur {
		case ' ':
			sc.advance()
		default:
			word := sc.scanWord()
			res = append(res, word)
		}
	}

	return res
}

func (sc *Scanner) scanWord() string {
	var (
		sb            strings.Builder
		isSingleQuote bool
		isDoubleQuote bool
		isEscaped     bool
	)

	for sc.cur != 0 {

		if isEscaped {
			isEscaped = false
			if isDoubleQuote && sc.cur != '"' && sc.cur != '\\' {
				sb.WriteRune('\\')
				sb.WriteRune(sc.cur)
			} else {
				sb.WriteRune(sc.cur)
			}
			sc.advance()
			continue
		}

		if sc.cur == '\\' && !isSingleQuote {
			isEscaped = true
			sc.advance()
			continue
		}

		if sc.cur == '\'' && !isDoubleQuote {
			isSingleQuote = !isSingleQuote
			sc.advance()
			continue
		}

		if sc.cur == '"' && !isSingleQuote {
			isDoubleQuote = !isDoubleQuote
			sc.advance()
			continue
		}

		if !isSingleQuote && !isDoubleQuote {
			if sc.cur == ' ' {
				break
			}
		}

		sb.WriteRune(sc.cur)
		sc.advance()
	}
	return sb.String()
}

func (sc *Scanner) advance() {
	sc.pos += 1

	if sc.pos >= len(sc.input) {
		// reach end, reset to rune=0
		sc.cur = 0
	} else {
		sc.cur = rune(sc.input[sc.pos])
	}
}

func (sc *Scanner) peek() rune {
	if sc.pos+1 >= len(sc.input) {
		// reach end, returns 0
		return 0
	}
	return rune(sc.input[sc.pos+1])
}
