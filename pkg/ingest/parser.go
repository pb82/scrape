package ingest

import (
	"errors"
	"fmt"
	"scrape/api"
	"strconv"
)

// Parser parses metrics responses in Prometheus format
type Parser struct {
	index  int
	tokens TokenList
}

func NewParser(tokens TokenList) *Parser {
	return &Parser{
		index:  0,
		tokens: tokens,
	}
}

func (p *Parser) hasTokens() bool {
	return p.index < len(p.tokens)
}

func (p *Parser) consume() {
	p.index = p.index + 1
}

func (p *Parser) next() (*Token, error) {
	if !p.hasTokens() {
		return nil, errors.New("unexpected end of stream")
	}
	current := p.index
	p.index = p.index + 1
	return p.tokens.at(current), nil
}

func (p *Parser) peek() (*Token, error) {
	if !p.hasTokens() {
		return nil, errors.New("unexpected end of stream")
	}
	return p.tokens.at(p.index), nil
}

func (p *Parser) expect(t TokenType) (*Token, error) {
	token, err := p.next()
	if err != nil {
		return nil, err
	}

	if token.TokenType == t {
		return token, nil
	}

	return nil, errors.New(fmt.Sprintf("unexpected token, expected %v but got %v:%v in line %v", TokenMapping[t], TokenMapping[token.TokenType], token.StringVal, token.Line))
}

func (p *Parser) label() (*api.Label, error) {
	label := &api.Label{}

	name, err := p.expect(TokenTypeName)
	if err != nil {
		return label, err
	}

	_, err = p.expect(TokenTypeEquals)
	if err != nil {
		return label, err
	}

	_, err = p.expect(TokenTypeQuote)
	if err != nil {
		return label, err
	}

	la, err := p.peek()
	if err != nil {
		return label, err
	}

	labelValue := ""
	if la.TokenType != TokenTypeQuote {
		value, err := p.expect(TokenTypeName)
		if err != nil {
			return label, err
		}
		labelValue = value.StringVal
	}

	_, err = p.expect(TokenTypeQuote)
	if err != nil {
		return label, err
	}

	label.Name = name.StringVal
	label.Value = labelValue
	return label, nil
}

func (p *Parser) labels() ([]api.Label, error) {
	var labels []api.Label
	for p.hasTokens() {
		label, err := p.label()
		if err != nil {
			return nil, err
		}
		labels = append(labels, *label)

		la, err := p.peek()
		if err != nil {
			return nil, err
		}
		if la.TokenType == TokenTypeComma {
			p.consume()
			continue
		} else if la.TokenType == TokenTypeRBrace {
			break
		} else {
			return nil, errors.New(fmt.Sprintf("unexpected token: expected , or } but got: %v", TokenMapping[la.TokenType]))
		}
	}
	return labels, nil
}

func (p *Parser) timeseries() (*api.Sample, error) {
	// <metric>{<label>="<value>", ...} <value>
	sample := &api.Sample{}
	token, err := p.expect(TokenTypeName)
	if err != nil {
		return nil, err
	}

	// assign the name label
	sample.Labels = append(sample.Labels, api.Label{
		Name:  "__name__",
		Value: token.StringVal,
	})

	la, err := p.peek()
	if err != nil {
		return nil, err
	}

	if la.TokenType == TokenTypeLBrace {
		// label list
		p.consume()
		labels, err := p.labels()
		if err != nil {
			return nil, err
		}
		sample.Labels = append(sample.Labels, labels...)
		_, err = p.expect(TokenTypeRBrace)
		if err != nil {
			return nil, err
		}
	}

	// sample value
	sampleValue, err := p.expect(TokenTypeName)
	if err != nil {
		return sample, err
	}

	var parsedValue = 0.0
	if sampleValue.StringVal != "" {
		parsedValue, err = strconv.ParseFloat(sampleValue.StringVal, 64)
		if err != nil {
			return sample, err
		}
	}

	sample.Value = parsedValue

	return sample, nil
}

func (p *Parser) Parse() ([]api.Sample, error) {
	var samples []api.Sample

	for p.hasTokens() {
		sample, err := p.timeseries()
		if err != nil {
			return nil, err
		}

		samples = append(samples, *sample)
	}

	return samples, nil
}
