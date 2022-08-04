package processing

import (
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func Metrics(expressions []string) map[string][]string {
	metrics := make(map[string][]string)
	for _, expr := range expressions {
		metricsInExpr := parser.Metrics(expr)
		for _, metric := range metricsInExpr {
			if _, ok := metrics[metric]; !ok {
				metrics[metric] = nil
			}
		}
	}
	return metrics
}

func MetricIdentifiers(metric string, metricMetadata []v1.MetricMetadata) map[string]struct{} {
	identifiers := make(map[string]struct{})
	for _, metadata := range metricMetadata {
		identifier := generateIdentifier(metadata.Target[Namespace], metadata.Target[Job])
		identifiers[identifier] = struct{}{}
	}
	return identifiers
}
