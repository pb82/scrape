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
				},
				{
					TokenType: TokenTypeValue,
					StringVal: "1.0",
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
				},
				{
					TokenType: TokenTypeLBrace,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
				},
				{
					TokenType: TokenTypeEquals,
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeRBrace,
				},
				{
					TokenType: TokenTypeValue,
					StringVal: "1.0",
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
				},
				{
					TokenType: TokenTypeLBrace,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
				},
				{
					TokenType: TokenTypeEquals,
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeComma,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "foo",
				},
				{
					TokenType: TokenTypeEquals,
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "bar",
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeRBrace,
				},
				{
					TokenType: TokenTypeValue,
					StringVal: "1.0",
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
				},
				{
					TokenType: TokenTypeLBrace,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "name",
				},
				{
					TokenType: TokenTypeEquals,
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "value",
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeComma,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "foo",
				},
				{
					TokenType: TokenTypeEquals,
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeName,
					StringVal: "bar",
				},
				{
					TokenType: TokenTypeQuote,
				},
				{
					TokenType: TokenTypeRBrace,
				},
				{
					TokenType: TokenTypeValue,
					StringVal: "1.0",
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
