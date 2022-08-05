package prom

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func GetRules(pathToRulesFile string) v1.RulesResult {
	if pathToRulesFile != "" {
		return getRulesFromFile(pathToRulesFile)
	}
	return Rules()
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
