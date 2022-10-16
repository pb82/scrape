package promql

import (
	"database/sql"
	"errors"
	"time"
)

type PromQlASTElement interface {
	Eval(db *sql.DB) error
}

type PromQlTimeseries struct {
	Name     string
	Duration time.Duration
}

type PromQlParser struct {
	index  int
	tokens []PromQlToken
}

type AggregatedLabel struct {
	LabelName  string
	LabelValue string
}

type AggregatedTimeseries struct {
	Metric string
	Labels []AggregatedLabel
	Value  float64
}

func (s *PromQlTimeseries) Eval(db *sql.DB) error {
	// get id of name label
	stmt, err := db.Prepare(`select id from labels where name = ?;`)
	if err != nil {
		return err
	}

	result := stmt.QueryRow("__name__")
	if result.Err() != nil {
		return result.Err()
	}

	var id int64
	err = result.Scan(&id)
	if err != nil {
		return err
	}

	// get samples id
	stmt, err = db.Prepare(`
select s.timeseries_id, tl.label_id, tl.label_value, s.value  from samples as s join timeseries_labels as tl on s.timeseries_id = tl.timeseries_id where s.timeseries_id in (select timeseries_id from timeseries_labels where label_id = ? and label_value = ?);
`)
	if err != nil {
		return err
	}

	timeseries, err := stmt.Query(id, s.Name)
	if result.Err() != nil {
		return result.Err()
	}

	type sample = struct {
		timeseries_id int
		label_id      int
		label_value   string
		sample_value  float64
	}

	var samples []sample
	for timeseries.Next() {
		var s sample
		err = timeseries.Scan(&s.timeseries_id, &s.label_id, &s.label_value, &s.sample_value)
		if err != nil {
			return err
		}
		samples = append(samples, s)
	}

	var aggregated map[int]AggregatedTimeseries
	for _, sample := range samples {
		if _, ok := aggregated[sample.timeseries_id]; !ok {
			aggregated[sample.timeseries_id] = AggregatedTimeseries{}
		}
	}

	return nil
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
	t := &PromQlTimeseries{}
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
