package parser

import "github.com/prometheus/prometheus/promql/parser"

// Metrics takes a PromQL expression and extracts all the metrics in that
// expression to a slice of strings. Duplicates will only show up once.
func Metrics(expression string) []string {
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
