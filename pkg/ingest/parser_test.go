package ingest

import "testing"

func Test_Parser(t *testing.T) {
	timeseries := `
# a comment
memory_usage{pod="a", namespace="n"} 512.0
memory_usage{pod="b", namespace="n"} 512.0
memory_usage{pod="c", namespace="m"} 512.0
# latency
latency 1.0
# alive
up{pod="a"} 1.0
`
	scanner := NewScanner()
	tokens, err := scanner.Scan(timeseries)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(tokens)
	_, err = parser.Parse()
	if err != nil {
		t.Fatal(err)
	}
}
