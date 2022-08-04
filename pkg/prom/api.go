package prom

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

const failureExitCode = -1

var promAPI v1.API

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
	client, err := api.NewClient(api.Config{
		Address:      url.String(),
		RoundTripper: roundTripper,
	})
	if err != nil {
		return fmt.Errorf("error creating API client: %s", err)
	}
	promAPI = v1.NewAPI(client)


	return nil
}

func GetJobsThatExportMetric(metric string) map[string]struct{} {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	metricsMetadata, err := promAPI.TargetsMetadata(ctx, "", metric, "")

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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	targets, err := promAPI.Targets(ctx)

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

func ValidateTime(s string) error {
	_, err := parseTime(s)
	return err
}
