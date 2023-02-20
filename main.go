package main

import (
	"flag"
	"fmt"
	"log"
	"sort"

	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/processing"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/prom"
)

var (
	cmd                *string
	server             *string
	metricsFile        *string
	exprFile           *string
	rulesFile          *string
	bearertoken        *string
	start              *string
	end                *string
	nonMatchingMetrics *bool
)

func init() {
	cmd = flag.String("cmd", "rules", `Options are:
  - metric-nb-series: the script will query rules/alerts from prometheus or a rules file and print for each metrics how many series exist in the TSDB.
  - rules: (default) the script will query rules/alerts from prometheus or a rules file and for each ServiceMonitor it will print which metrics exposes.
  - expressions-file: the script will read the file where each line is a PromQL expression and for each ServiceMonitor it will print which metrics exposes.
  - metrics-file: the script will read the file where each line is a metric and for each ServiceMonitor it will print which metrics exposes.
  - all: takes as source both the expression file and rules/alerts from prometheus or a rules file and or each ServiceMonitor it will print which metrics exposes.`)
	server = flag.String("server", "http://localhost:9090/", "Prometheus server URL")
	metricsFile = flag.String("metrics-file", "", "File with list of metrics")
	exprFile = flag.String("expressions-file", "", "File with list of PromQL expressions")
	rulesFile = flag.String("rules-file", "", "File with rules file in json format, usualy taken from a mustgather")
	bearertoken = flag.String("bearer-token", "", "Bearer Token to connect to the server")
	start = flag.String("start", "", "Start time (RFC3339 or Unix timestamp).")
	end = flag.String("end", "", "End time (RFC3339 or Unix timestamp).")
	nonMatchingMetrics = flag.Bool("non-matching-metrics", false, "Return metrics do not match a Service Monitor")
}

func main() {

	flag.Parse()
	validateArguments()
	var (
		metrics []string
		err     error
	)

	switch {
	case *cmd == "metric-nb-series":
		printMetricsAndNumberOfSeries(*rulesFile)
	case *cmd == "expressions-file":
		if *exprFile == "" {
			log.Fatal("error the \"expressions-file\" command requires the flag \"--expressions-file\" to be passed")
		}
		expressions, err := parser.ReadLines(*exprFile)
		if err != nil {
			log.Fatalf("error while reading expressions from file: %s", err)
		}
		for metric := range processing.Metrics(expressions) {
			metrics = append(metrics, metric)
		}
	case *cmd == "metrics-file":
		if *metricsFile == "" {
			log.Fatal("error the \"metrics-file\" command requires the flag \"--metrics-file\" to be passed")
		}
		metrics, err = parser.ReadLines(*metricsFile)
		if err != nil {
			log.Fatalf("error while reading metrics from file: %s", err)
		}
	case *cmd == "all":
		if *exprFile == "" {
			log.Fatal("error the \"all\" command requires the flag \"--expressions-file\" to be passed")
		}
		expressions, err := parser.ReadLines(*exprFile)
		if err != nil {
			log.Fatalf("error while reading expressions from file: %s", err)
		}
		for metric := range processing.Metrics(expressions) {
			metrics = append(metrics, metric)
		}

		metrics = append(metrics, metricsFromRules(*rulesFile)...)
	case *cmd == "rules":
		metrics = metricsFromRules(*rulesFile)
	default:
		log.Fatalf("error the command passed \"%s\" does not match any of the supported commands, run the script with \"--h\" to see the list of supported commands", *cmd)
	}

	smMetrics := metricsPerServiceMonitor(metrics)
	fmt.Println("Metrics Per ServiceMonitor")
	printMapInOrder(smMetrics)
	if *nonMatchingMetrics {
		printMetricsNotExportedBySM(*rulesFile, metrics)
	}
}

func validateArguments() {
	if err := prom.SetUpClient(*server, *bearertoken); err != nil {
		log.Fatalf("error could not set up client: %s", err)
	}

	if err := prom.ValidateTime(*start); *start != "" && err != nil {
		log.Fatalf("error parameter start %s", err)
	}

	if err := prom.ValidateTime(*end); *end != "" && err != nil {
		log.Fatalf("error parameter end %s", err)
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
