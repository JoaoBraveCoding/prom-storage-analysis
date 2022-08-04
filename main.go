package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/processing"
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

	// By Service Monitor Command
	if *byServiceMonitor {
		printMetricsPerMetricGenerator(*pathToRulesFile)
		return
	}
	// Regular Command (Metrics & Nb of Series)
	printMetricsAndNumberOfSeries(*pathToRulesFile)
}

func validateArguments() {
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

func printMetricsAndNumberOfSeries(pathToRules string) {
	expressions := processing.Expressions(prom.GetRules(pathToRules))
	metrics := processing.Metrics(expressions)
	for _, metric := range toSortedArray(metrics) {
		fmt.Printf("%s %d\n", metric, prom.Series(metric, *start, *end))
	}
}

func printMetricsPerMetricGenerator(pathToRules string) {
	// Build a map with metrics and possible identifiers.
	// An identifier is the concatenation of namespace + / + job values
	// that we get for each metric using the Targets Metadata endpoint
	expressions := processing.Expressions(prom.GetRules(pathToRules))
	metricsIdentifiers := make(map[string]map[string]struct{})
	for metric := range processing.Metrics(expressions) {
		metricMetadata := prom.MetricMetadata(metric)
		metricsIdentifiers[metric] = processing.MetricIdentifiers(metric, metricMetadata)
	}

	// Compute and Print Metrics Per ServiceMonitor
	targets := prom.Targets()
	scrapeConfigsIdentifiers := processing.ScrapeConfigsIdentifiers(targets)
	scrapeConfigsMetrics := processing.ScrapeConfigsMetrics(scrapeConfigsIdentifiers, metricsIdentifiers)
	fmt.Println("Metrics Per ServiceMonitor")
	printMapInOrder(scrapeConfigsMetrics)

	// Not all metrics used in rules are exported by Monitors
	// Compute and Print Metrics Per PrometheusRule (recording rules)
	promRules := prom.GetRules(pathToRules)
	promRulesRecordingRules := processing.PromRulesRecordingRules(promRules)
	promRuleMetrics := processing.PromRuleMetrics(promRulesRecordingRules, metricsIdentifiers)
	fmt.Println("Metrics (recording rules) Per PrometheusRule")
	printMapInOrder(promRuleMetrics)

	// Metrics not anywhere
	fmt.Println("Metrics not exported by Service Monitor nor by PrometheusRule")
	for metric, identifiers := range metricsIdentifiers {
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

func printMapInOrder(m map[string][]string) {
	for _, val := range toSortedArray(m) {
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
