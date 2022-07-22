package main

import (
	"fmt"
	"net/url"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/prom"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	var expressions []string
	url := &url.URL{
		Host:   "localhost:9090",
		Scheme: "http",
		Path:   "/",
	}
	promRules := prom.GetRules(url)

	for _, group := range promRules.Groups {
		for _, r := range group.Rules {
			switch v := r.(type) {
			case v1.RecordingRule:
				expressions = append(expressions, v.Query)
			case v1.AlertingRule:
				expressions = append(expressions, v.Query)
			default:
				return
			}
		}
	}

	mapOfMetrics := make(map[string]bool)
	for _, expr := range expressions {
		metricsInExpr := parser.GetMetrics(expr)
		for _, metric := range metricsInExpr {
			if _, ok := mapOfMetrics[metric]; !ok {
				mapOfMetrics[metric] = true
			}
		}
	}

	for metric := range mapOfMetrics {
		fmt.Printf("%s %d\n", metric, prom.SeriesPerMetric(url, metric, "", ""))
	}
}
