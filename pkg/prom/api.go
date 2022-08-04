package prom

import (
	"context"
	"errors"
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

func Series(matcher string, start, end string) int {
	stime, etime, err := parseStartTimeAndEndTime(start, end)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failureExitCode
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	val, _, err := promAPI.Series(ctx, []string{matcher}, stime, etime) // Ignoring warnings for now.

	cancel()
	if err != nil {
		return handleAPIError(err)
	}

	return len(val)
}

func Rules() v1.RulesResult {
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

func MetricMetadata(metric string) []v1.MetricMetadata {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	metricMetadata, err := promAPI.TargetsMetadata(ctx, "", metric, "")

	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when preforming get to TargetMetadata endpoint:", err)
		return []v1.MetricMetadata{}
	}

	return metricMetadata
}

func Targets() v1.TargetsResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	targets, err := promAPI.Targets(ctx)

	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error when preforming get to Targets endpoint:", err)
		return v1.TargetsResult{}
	}

	return targets
}

func ValidateTime(s string) error {
	_, err := parseTime(s)
	return err
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
