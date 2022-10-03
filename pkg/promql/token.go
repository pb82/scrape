package promql

type PromQlTokenType string

const (
	PromQlName     PromQlTokenType = "name"
	PromQlLBracket PromQlTokenType = "["
	PromQlRBracket PromQlTokenType = "]"
)

type PromQlToken struct {
	Type  PromQlTokenType
	Value string
}
