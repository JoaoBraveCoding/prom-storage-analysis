package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/prom"
)

var (
	server *string
)

func init() {
	server = flag.String("url", "http://localhost:9090/", "server")
}

func main() {

	flag.Parse()

	parsedURL, err := url.Parse(*server)
	if err != nil {
		fmt.Println(fmt.Errorf("error while parsing server variable, %s", err))
		os.Exit(1)
	}

	expressions := prom.GetUsedExprInRules(parsedURL)

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
		fmt.Printf("%s %d\n", metric, prom.SeriesPerMetric(parsedURL, metric, "", ""))
	}
}
