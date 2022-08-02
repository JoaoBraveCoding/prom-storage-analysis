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
	server           *string
	pathToRulesFile  *string
	bearertoken      *string
	start            *string
	end              *string
	byServiceMonitor *bool
)

func init() {
	server = flag.String("server", "http://localhost:9090/", "Prometheus server URL")
	pathToRulesFile = flag.String("rules-file", "", "Path to a rules file in json from a mustgather")
	bearertoken = flag.String("bearer-token", "", "Bearer Token to connect to the server")
	start = flag.String("start", "", "Start time (RFC3339 or Unix timestamp).")
	end = flag.String("end", "", "End time (RFC3339 or Unix timestamp).")
	byServiceMonitor = flag.Bool("by-service-monitor", false, "Return metrics ordered by Service Monitor")
}

func main() {

	flag.Parse()

	if err := prom.SetUpClient(*server, *bearertoken); err != nil {
		log.Fatalf("error could not set up client: %s", err)
		os.Exit(1)
	}

	if err := prom.ValidateTime(*start); *start != "" && err != nil {
		log.Fatalf("error parameter start %s", err)
		os.Exit(1)
	}

	if err := prom.ValidateTime(*end); *end != "" && err != nil {
		log.Fatalf("error parameter end %s", err)
		os.Exit(1)
	}

	expressions := prom.GetUsedExprInRules(*pathToRulesFile)

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

	if !*byServiceMonitor {
		for _, metric := range sortedMetrics {
			fmt.Printf("%s %d\n", metric, prom.SeriesPerMetric(metric, *start, *end))
		}
	} else {
		// Build map with metrics and possible identifiers
		metricsIdentifier := make(map[string]map[string]struct{})
		for _, metric := range sortedMetrics {
			metricsIdentifier[metric] = prom.GetJobsThatExportMetric(metric)
		}

		// Build map with ScrapeConfig and possible identifiers
		scrapeConfigIdentifiers := prom.GetIdentifierPerScrapeConfig()

		// Build map with ScrapeConfig and metrics which match one of his identifiers
		scrapeConfigMetrics := make(map[string][]string)
		for scrapeConfig, scIdentifiers := range scrapeConfigIdentifiers {
			var metricsExposed []string
			for scIdentifier, _ := range scIdentifiers {
				for metric, mIdentifier := range metricsIdentifier {
					if _, ok := mIdentifier[scIdentifier]; ok {
						metricsExposed = append(metricsExposed, metric)
					}
				}
			}
			scrapeConfigMetrics[scrapeConfig] = metricsExposed
		}

		//Sorting
		sortedScrapeConfig := make([]string, len(scrapeConfigMetrics))
		i := 0
		for scrapeConfig := range scrapeConfigMetrics {
			sortedScrapeConfig[i] = scrapeConfig
			i++
		}
		sort.Strings(sortedScrapeConfig)

		//Printing
		for _, scrapeConfig := range sortedScrapeConfig {
			fmt.Println(scrapeConfig)
			sort.Strings(scrapeConfigMetrics[scrapeConfig])
			for _, metric := range scrapeConfigMetrics[scrapeConfig] {
				fmt.Println(metric)
			}
			fmt.Println("")
		}

		fmt.Println("Metrics without a Job/Namespace Label")
		for metric, identifiers := range metricsIdentifier {
			if len(identifiers) == 0 {
				fmt.Println(metric)
			}
		}
	}
}
