package processing

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func Expressions(promRules v1.RulesResult) (expressions []string) {
	for _, group := range promRules.Groups {
		for _, r := range group.Rules {
			switch v := r.(type) {
			case v1.RecordingRule:
				expressions = append(expressions, v.Query)
			case v1.AlertingRule:
				expressions = append(expressions, v.Query)
			default:
				fmt.Fprintln(os.Stderr, "error when parsing rules found rule which is not an AlertingRule nor a RecordingRule")
				os.Exit(1)
			}
		}
	}
	return expressions
}

func PromRulesRecordingRules(promRules v1.RulesResult) map[string]map[string]struct{} {
	promRulesRecordingRules := make(map[string]map[string]struct{})
	for _, group := range promRules.Groups {
		promRule := strings.Split(group.File, "/")[len(strings.Split(group.File, "/"))-1]
		if _, exists := promRulesRecordingRules[promRule]; !exists {
			promRulesRecordingRules[promRule] = make(map[string]struct{})
		}
		for _, r := range group.Rules {
			switch v := r.(type) {
			case v1.RecordingRule:
				promRulesRecordingRules[promRule][v.Name] = struct{}{}
			case v1.AlertingRule:
			default:
				fmt.Fprintln(os.Stderr, "error when parsing rules found rule which is not an AlertingRule nor a RecordingRule")
				os.Exit(1)
			}
		}
	}
	return promRulesRecordingRules
}

func PromRuleMetrics(promRulesRecordingRules, metricsIdentifiers map[string]map[string]struct{}) map[string][]string {
	promRulesMetrics := make(map[string][]string)
	for promRule, recordingRules := range promRulesRecordingRules {
		promRulesMetrics[promRule] = make([]string, 0)
		for recordingRule := range recordingRules {
			if identifiers, ok := metricsIdentifiers[recordingRule]; ok {
				if len(identifiers) == 0 {
					promRulesMetrics[promRule] = append(promRulesMetrics[promRule], recordingRule)
				}
			}
		}
	}
	return promRulesMetrics
}
