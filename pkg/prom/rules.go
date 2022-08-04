package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func GetExprUsedInRules(pathToRulesFile string) (expressions []string) {
	promRules := getRules(pathToRulesFile)

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

func GetRecodringRulesPerPromRule(pathToRulesFile string) map[string][]string {
	promRules := getRules(pathToRulesFile)

	recordingRulesPerPromRule := make(map[string][]string)
	for _, group := range promRules.Groups {
		recordingRulesPerPromRule[group.Name] = make([]string, 0)
		for _, r := range group.Rules {
			switch v := r.(type) {
			case v1.RecordingRule:
				recordingRulesPerPromRule[group.Name] = append(recordingRulesPerPromRule[group.Name], v.Name)
			case v1.AlertingRule:
				break
			default:
				fmt.Fprintln(os.Stderr, "error when parsing rules found rule which is not an AlertingRule nor a RecordingRule")
				os.Exit(1)
			}
		}
	}
	return recordingRulesPerPromRule
}

func getRules(pathToRulesFile string) v1.RulesResult {
	if pathToRulesFile != "" {
		return getRulesFromFile(pathToRulesFile)
	}
	return getRulesFromProm()
}

func getRulesFromProm() v1.RulesResult {
	// Run query against client.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	rules, err := promAPI.Rules(ctx)

	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when preforming get to Rules endpoint:", err)
		return v1.RulesResult{}
	}

	return rules
}

func getRulesFromFile(pathToRulesFile string) v1.RulesResult {
	jsonContent, err := os.Open(pathToRulesFile)
	if err != nil {
		log.Fatalf("error opening rules file: %s", err)
	}
	defer jsonContent.Close()

	byteValue, _ := ioutil.ReadAll(jsonContent)

	type data struct {
		Groups []v1.RuleGroup
	}

	type rulesFile struct {
		Status string
		Data   data
	}
	var rules rulesFile

	json.Unmarshal(byteValue, &rules)

	return v1.RulesResult{Groups: rules.Data.Groups}
}
