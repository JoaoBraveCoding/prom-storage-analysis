package parser

import "github.com/prometheus/prometheus/promql/parser"

// Extracts metrics used a PromQL expression
// duplicates are removed
func GetMetrics(expression string) []string {
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		panic(err)
	}

	metrics := make(map[string]bool)
	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		switch n := node.(type) {
		case *parser.VectorSelector:
			metrics[n.Name] = true
		}
		return nil
	})

	usedMetrics := make([]string, len(metrics))
	i := 0
	for metric := range metrics {
		usedMetrics[i] = metric
		i++
	}
	return usedMetrics
}
