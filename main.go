package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/prom"
)

var (
	server      *string
	bearertoken *string
	start       *string
	end         *string
)

func init() {
	server = flag.String("url", "http://localhost:9090/", "server")
	bearertoken = flag.String("bearer-token", "", "Bearer Token to connect to the server")
	start = flag.String("start", "", "Start time (RFC3339 or Unix timestamp).")
	end = flag.String("end", "", "End time (RFC3339 or Unix timestamp).")
}

func main() {

	flag.Parse()

	if err := prom.SetUpClient(*server, *bearertoken); err != nil {
		log.Fatalf("error could not set up client: %s", err)
		os.Exit(1)
	}

	if err := prom.ValidateTime(*start); *start != "" && err != nil{
		log.Fatalf("error parameter start %s", err)
		os.Exit(1)		
	}

	if err := prom.ValidateTime(*end); *end != "" && err != nil{
		log.Fatalf("error parameter end %s", err)
		os.Exit(1)		
	}

	expressions := prom.GetUsedExprInRules()

	mapOfMetrics := make(map[string]struct{})
	for _, expr := range expressions {
		metricsInExpr := parser.GetMetrics(expr)
		for _, metric := range metricsInExpr {
			if _, ok := mapOfMetrics[metric]; !ok {
				mapOfMetrics[metric] = struct{}{}
			}
		}
	}

	sortedMetrics := make([]string, len(mapOfMetrics))
	i := 0
	for metric := range mapOfMetrics {
		sortedMetrics[i] = metric
		i++
	}
	sort.Strings(sortedMetrics)

	for _, metric := range sortedMetrics {
		fmt.Printf("%s %d\n", metric, prom.SeriesPerMetric(metric, *start, *end))
	}
}
