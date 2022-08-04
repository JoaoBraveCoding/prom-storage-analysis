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
	validateArguments()
	
	mapOfMetrics := getUsedMetricsInRules()
	// By Service Monitor Command
	if *byServiceMonitor {
		printMetricsPerMetricGenerator(mapOfMetrics)
		return
	}
	// Regular Command (Metrics & Nb of Series)
	printMetricsAndNumberOfSeries(mapOfMetrics)
}

func validateArguments(){
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
}

func getUsedMetricsInRules() map[string][]string {
	expressions := prom.GetExprUsedInRules(*pathToRulesFile)

	mapOfMetrics := make(map[string][]string)
	for _, expr := range expressions {
		metricsInExpr := parser.GetMetrics(expr)
		for _, metric := range metricsInExpr {
			if _, ok := mapOfMetrics[metric]; !ok {
				mapOfMetrics[metric] = nil
			}
		}
	}

	return mapOfMetrics
}

func printMetricsAndNumberOfSeries(mapOfMetrics map[string][]string) {
	sortedMetrics := toSortedArray(mapOfMetrics)
	for _, metric := range sortedMetrics {
		fmt.Printf("%s %d\n", metric, prom.SeriesPerMetric(metric, *start, *end))
	}
}

func printMetricsPerMetricGenerator(mapOfMetrics map[string][]string) {
	// Build a map with metrics and possible identifiers.
	// An identifier is the concatenaiton of namespace + / + job values
	// that we get for each metric using the Targets Metadata endpoint
	metricsIdentifier := make(map[string]map[string]struct{})
	for metric := range mapOfMetrics {
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
	sortedScrapeConfig := toSortedArray(scrapeConfigMetrics)

	//Printing
	fmt.Println("Metrics Per ServiceMonitor")
	printMapInOrder(scrapeConfigMetrics, sortedScrapeConfig)

	recordingRulePerPromRule := prom.GetRecodringRulesPerPromRule(*pathToRulesFile)
	promRuleMetrics := make(map[string][]string)
	for promRule, recordingRules := range recordingRulePerPromRule {
		promRuleMetrics[promRule] = make([]string, 0)
		for _, recordingRule := range recordingRules {
			if identifiers, ok := metricsIdentifier[recordingRule]; ok {
				if len(identifiers) == 0 {
					promRuleMetrics[promRule] = append(promRuleMetrics[promRule], recordingRule)
				}
			}
		}
	}

	//Sorting
	sortedPromRule := toSortedArray(promRuleMetrics)

	//Printing
	fmt.Println("Metrics (recording rules) Per PrometheusRule")
	printMapInOrder(promRuleMetrics, sortedPromRule)

	// Metrics not anywhere
	fmt.Println("Metrics not exported by Service Monitor nor by PrometheusRule")
	for metric, identifiers := range metricsIdentifier {
		if len(identifiers) == 0 {
			found := false
			for _, recordingRules := range promRuleMetrics {
				for _, recordingRule := range recordingRules {
					if recordingRule == metric {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				fmt.Println(metric)
			}
		}
	}

}

func toSortedArray(m map[string][]string) []string {
	sortedMap := make([]string, len(m))
	i := 0
	for promRule := range m {
		sortedMap[i] = promRule
		i++
	}
	sort.Strings(sortedMap)
	return sortedMap
}

func printMapInOrder(m map[string][]string, order []string) {
	for _, val := range order {
		if len(m[val]) == 0 {
			continue
		}
		fmt.Println(val)
		sort.Strings(m[val])
		for _, jVal := range m[val] {
			fmt.Println(jVal)
		}
	}
}
