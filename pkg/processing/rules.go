package processing

import (
	"fmt"
	"os"

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

func PromRulesRecordingRules(promRules v1.RulesResult) map[string][]string {
	promRulesRecordingRules := make(map[string][]string)
	for _, group := range promRules.Groups {
		promRulesRecordingRules[group.Name] = make([]string, 0)
		for _, r := range group.Rules {
			switch v := r.(type) {
			case v1.RecordingRule:
				promRulesRecordingRules[group.Name] = append(promRulesRecordingRules[group.Name], v.Name)
			case v1.AlertingRule:
				break
			default:
				fmt.Fprintln(os.Stderr, "error when parsing rules found rule which is not an AlertingRule nor a RecordingRule")
				os.Exit(1)
			}
		}
	}
	return promRulesRecordingRules
}

func PromRuleMetrics(promRulesRecordingRules map[string][]string, metricsIdentifiers map[string]map[string]struct{}) map[string][]string {
	promRuleMetrics := make(map[string][]string)
	for promRule, recordingRules := range promRulesRecordingRules {
		promRuleMetrics[promRule] = make([]string, 0)
		for _, recordingRule := range recordingRules {
			if identifiers, ok := metricsIdentifiers[recordingRule]; ok {
				if len(identifiers) == 0 {
					promRuleMetrics[promRule] = append(promRuleMetrics[promRule], recordingRule)
				}
			}
		}
	}
	return promRuleMetrics
}
