package api

type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Sample struct {
	Labels []Label `json:"labels"`
	Value  float64 `json:"value"`
}
