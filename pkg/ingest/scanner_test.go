package ingest

import (
	"reflect"
	"testing"
)

func Test_Scanner(t *testing.T) {
	type testcase struct {
		data       string
		wantError  bool
		wantTokens TokenList
	}

	testcases := []testcase{
		{
			data: `
# a comment
`,
			wantError:  false,
			wantTokens: nil,
		},
		{
			data: `
a_metric
`,
			wantError: false,
			wantTokens: TokenList{
				{
					TokenType: TokenTypeName,
					StringVal: "a_metric",
					Line:      1,
				},
			},
		},
		{
			data: `
a_metric 1.0
`,
			wantError: false,
			wantTokens: TokenList{
				{
					TokenType: TokenTypeName,
					StringVal: "a_metric",
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "1.0",
					Line:      1,
				},
			},
		},
		{
			data: `
a_metric{name="value"} 1.0
`,
			wantError: false,
			wantTokens: TokenList{
				{
					TokenType: TokenTypeName,
					StringVal: "a_metric",
					Line:      1,
				},
				{
					TokenType: TokenTypeLBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
					Line:      1,
				},
				{
					TokenType: TokenTypeEquals,
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeRBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "1.0",
					Line:      1,
				},
			},
		},
		{
			data: `
a_metric{name="value", foo="bar"} 1.0
`,
			wantError: false,
			wantTokens: TokenList{
				{
					TokenType: TokenTypeName,
					StringVal: "a_metric",
					Line:      1,
				},
				{
					TokenType: TokenTypeLBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
					Line:      1,
				},
				{
					TokenType: TokenTypeEquals,
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeComma,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "foo",
					Line:      1,
				},
				{
					TokenType: TokenTypeEquals,
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "bar",
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeRBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "1.0",
					Line:      1,
				},
			},
		},
		{
			data: `
a_metric{name="value", foo="bar"} 1.0
`,
			wantError: false,
			wantTokens: TokenList{
				{
					TokenType: TokenTypeName,
					StringVal: "a_metric",
					Line:      1,
				},
				{
					TokenType: TokenTypeLBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
					Line:      1,
				},
				{
					TokenType: TokenTypeEquals,
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeComma,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "foo",
					Line:      1,
				},
				{
					TokenType: TokenTypeEquals,
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "bar",
					Line:      1,
				},
				{
					TokenType: TokenTypeQuote,
					Line:      1,
				},
				{
					TokenType: TokenTypeRBrace,
					Line:      1,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "1.0",
					Line:      1,
				},
			},
		},
	}

	for _, tc := range testcases {
		scanner := NewScanner()
		tokens, err := scanner.Scan(tc.data)

		if !tc.wantError && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(tc.wantTokens, tokens) {
			t.Fatalf("unexpected result, wanted %v, got %v", tc.wantTokens, tokens)
		}
	}
}
