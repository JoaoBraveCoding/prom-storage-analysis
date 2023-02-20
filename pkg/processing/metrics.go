package processing

import (
	"strings"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/prom"
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

// BuilMetricsIdentifiers build a map with metrics and possible identifiers.
// An identifier is the concatenation of the values namespace + / + job
// that we get for each metric using the Targets Metadata endpoint
func BuilMetricsIdentifiers(metrics []string) map[string]map[string]struct{} {
	metricsIdentifiers := make(map[string]map[string]struct{})
	for _, metric := range metrics {
		if metric == "" {
			continue
		}
		metricMetadata := prom.MetricMetadata(metric)
		if len(metricMetadata) == 0 {
			// Lookup for the parent metric name if it's a counter, histogram or summary.
			for _, s := range []string{"_total", "_bucket", "_sum", "_count"} {
				metricMetadata = prom.MetricMetadata(strings.TrimSuffix(metric, s))
				if len(metricMetadata) > 0 {
					break
				}
			}
		}
		metricsIdentifiers[metric] = MetricIdentifiers(metric, metricMetadata)
	}
	return metricsIdentifiers
}
