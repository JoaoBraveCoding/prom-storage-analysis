package prom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

const failureExitCode = -1

var client api.Client

func SetUpClient(server, bearertoken string) error {
	url, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("error while parsing server variable, %s", err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}

	roundTripper := api.DefaultRoundTripper
	if bearertoken != "" {
		roundTripper = config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(bearertoken), api.DefaultRoundTripper)
	}

	// Create new client.
	client, err = api.NewClient(api.Config{
		Address:      url.String(),
		RoundTripper: roundTripper,
	})
	if err != nil {
		return fmt.Errorf("error creating API client: %s", err)
	}

	return nil
}

func GetUsedExprInRules(pathToRulesFile string) (expressions []string) {
	var promRules v1.RulesResult
	if pathToRulesFile == "" {
		promRules = getRules()
	} else {
		promRules = getRulesFromFile(pathToRulesFile)
	}

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

func getRules() v1.RulesResult {
	// Run query against client.
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	rules, err := api.Rules(ctx)

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

func SeriesPerMetric(matcher string, start, end string) int {
	stime, etime, err := parseStartTimeAndEndTime(start, end)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failureExitCode
	}

	// Run query against client.
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	val, _, err := api.Series(ctx, []string{matcher}, stime, etime) // Ignoring warnings for now.

	cancel()
	if err != nil {
		return handleAPIError(err)
	}

	return len(val)
}

func GetJobsThatExportMetric(metric string) map[string]struct{} {
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	metricsMetadata, err := api.TargetsMetadata(ctx, "", metric, "")

	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when preforming get to TargetMetadata endpoint:", err)
		return map[string]struct{}{}
	}

	jobsNamespaces := make(map[string]struct{})
	for _, metricMetadata := range metricsMetadata {
		identifier := metricMetadata.Target["namespace"] + "/" + metricMetadata.Target["job"]
		jobsNamespaces[identifier] = struct{}{}
	}
	return jobsNamespaces
}

func GetIdentifierPerScrapeConfig() map[string]map[string]struct{} {
	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	targets, err := api.Targets(ctx)

	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when preforming get to Targets endpoint:", err)
		return map[string]map[string]struct{}{}
	}

	scrapeConfigIdentifiers := make(map[string]map[string]struct{})
	for _, target := range targets.Active {
		splitedScrapePool := strings.Split(target.ScrapePool, "/")
		identifier := string(target.Labels["namespace"] + "/" + target.Labels["job"])
		scrapeConfig := splitedScrapePool[1] + "/" + splitedScrapePool[2]
		if identifiers, ok := scrapeConfigIdentifiers[scrapeConfig]; ok {
			if _, ok := identifiers[identifier]; !ok {
				identifiers[identifier] = struct{}{}
			}
		} else {
			identifiers := make(map[string]struct{})
			identifiers[identifier] = struct{}{}
			scrapeConfigIdentifiers[scrapeConfig] = identifiers
		}

	}

	return scrapeConfigIdentifiers
}

func parseStartTimeAndEndTime(start, end string) (time.Time, time.Time, error) {
	var (
		minTime = time.Now().Add(-2 * time.Hour)
		maxTime = time.Now().Add(2 * time.Hour)
		err     error
	)

	stime := minTime
	etime := maxTime

	if start != "" {
		stime, err = parseTime(start)
		if err != nil {
			return stime, etime, fmt.Errorf("error parsing start time: %w", err)
		}
	}

	if end != "" {
		etime, err = parseTime(end)
		if err != nil {
			return stime, etime, fmt.Errorf("error parsing end time: %w", err)
		}
	}
	return stime, etime, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := strconv.ParseFloat(s, 64); err == nil {
		s, ns := math.Modf(t)
		return time.Unix(int64(s), int64(ns*float64(time.Second))).UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q to a valid timestamp", s)
}

func handleAPIError(err error) int {
	var apiErr *v1.Error
	if errors.As(err, &apiErr) && apiErr.Detail != "" {
		fmt.Fprintf(os.Stderr, "query error: %v (detail: %s)\n", apiErr, strings.TrimSpace(apiErr.Detail))
	} else {
		fmt.Fprintln(os.Stderr, "query error:", err)
	}

	return failureExitCode
}

func ValidateTime(s string) error {
	_, err := parseTime(s)
	return err
}
