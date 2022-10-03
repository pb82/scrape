package promql

import "testing"

func Test_PromQlScanner(t *testing.T) {
	source := "grafana_feature_toggles_info[5m]"
	scanner := NewPromQlScanner(source)
	tokens := scanner.ScanPromQl()
	t.Log(tokens)
}
