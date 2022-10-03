package promql

import (
	"strings"
	"unicode"
)

type PromQlScanner struct {
	index int
	runes []rune
}

func NewPromQlScanner(source string) *PromQlScanner {
	return &PromQlScanner{
		index: 0,
		runes: []rune(source),
	}
}

func (s *PromQlScanner) peek() rune {
	return s.runes[s.index]
}

func (s *PromQlScanner) next() rune {
	cr := s.runes[s.index]
	s.index++
	return cr
}

func (s *PromQlScanner) consume() {
	s.index++
}

func (s *PromQlScanner) name() PromQlToken {
	sb := strings.Builder{}
	for s.index < len(s.runes) {
		p := s.peek()
		if unicode.IsSpace(p) {
			break
		}

		switch p {
		case '[', ']':
			goto end
			break
		default:
			s.consume()
			sb.WriteRune(p)
		}
	}
end:

	return PromQlToken{
		Type:  PromQlName,
		Value: sb.String(),
	}
}

func (s *PromQlScanner) ScanPromQl() []PromQlToken {
	var tokens []PromQlToken

	for s.index < len(s.runes) {
		r := s.peek()

		if unicode.IsSpace(r) {
			s.consume()
			continue
		}

		switch r {
		case '[':
			s.consume()
			tokens = append(tokens, PromQlToken{
				Type: PromQlLBracket,
			})
		case ']':
			s.consume()
			tokens = append(tokens, PromQlToken{
				Type: PromQlRBracket,
			})
		default:
			tokens = append(tokens, s.name())
		}
	}

	return tokens
}
