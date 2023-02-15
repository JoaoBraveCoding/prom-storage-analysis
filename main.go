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
		metrics := metricsFromRules(*pathToRulesFile)
		smMetrics := metricsPerServiceMonitor(metrics)
		fmt.Println("Metrics Per ServiceMonitor")
		printMapInOrder(smMetrics)
		printMetricsNotExportedBySM(*pathToRulesFile, metrics)
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

// metricsFromRules extract metrics from rules present in pathToRules, if path
// is "" then get rules from Prometheus on localhost
func metricsFromRules(pathToRules string) []string {
	metrics := make([]string, 0)
	expressions := processing.Expressions(prom.GetRules(pathToRules))
	for metric := range processing.Metrics(expressions) {
		metrics = append(metrics, metric)
	}
	return metrics
}

// metricsPerServiceMonitor generates map where key is ServiceMonitor
// and value is an array of metrics exposed by that ServiceMonitor 
func metricsPerServiceMonitor(metrics []string) map[string][]string {
	metricsIdentifiers := processing.BuilMetricsIdentifiers(metrics)

	// Compute and Print Metrics Per ServiceMonitor
	targets := prom.Targets()
	scrapeConfigsIdentifiers := processing.ScrapeConfigsIdentifiers(targets)
	return processing.ScrapeConfigsMetrics(scrapeConfigsIdentifiers, metricsIdentifiers)
}

// Not all metrics used in rules are exported by Monitors
// Compute and Print Metrics Per PrometheusRule (recording rules)
func printMetricsNotExportedBySM(pathToRules string, metrics []string) {
	metricsIdentifiers := processing.BuilMetricsIdentifiers(metrics)
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
			if !found && prom.Series(metric, *start, *end) != 0 {
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
