package promql

import (
	"errors"
	"time"
)

type PromQlASTElement interface {
	Eval()
}

type PromQlTimeseries struct {
	Name     string
	Duration time.Duration
}

type PromQlParser struct {
	index  int
	tokens []PromQlToken
}

func (s PromQlTimeseries) Eval() {

}

func NewPromQlParser(tokens []PromQlToken) *PromQlParser {
	return &PromQlParser{
		index:  0,
		tokens: tokens,
	}
}

func (s *PromQlParser) peek() *PromQlToken {
	if s.index >= len(s.tokens) {
		return nil
	}
	t := s.tokens[s.index]
	return &t
}

func (s *PromQlParser) next() (*PromQlToken, error) {
	if s.index >= len(s.tokens) {
		return nil, errors.New("expected token but reached end of stream")
	}
	t := s.tokens[s.index]
	s.index++
	return &t, nil
}

func (s *PromQlParser) expect(tokenType PromQlTokenType) (*PromQlToken, error) {
	t, err := s.next()
	if err != nil {
		return nil, err
	}

	if t.Type != tokenType {
		return nil, errors.New("unexpected token")
	}

	return t, nil
}

func (s *PromQlParser) timeseries() (PromQlASTElement, error) {
	t := PromQlTimeseries{}
	name, err := s.expect(PromQlName)
	if err != nil {
		return nil, err
	}
	t.Name = name.Value

	la := s.peek()
	if la != nil {
		_, err := s.expect(PromQlLBracket)
		if err != nil {
			return nil, err
		}

		duration, err := s.expect(PromQlName)
		if err != nil {
			return nil, err
		}

		_, err = s.expect(PromQlRBracket)
		if err != nil {
			return nil, err
		}

		d, err := time.ParseDuration(duration.Value)
		if err != nil {
			return nil, err
		}

		t.Duration = d
	}

	return t, nil
}

func (s *PromQlParser) Parse() (PromQlASTElement, error) {
	return s.timeseries()
}
