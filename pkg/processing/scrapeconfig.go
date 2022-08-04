package processing

import (
	"strings"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func ScrapeConfigsIdentifiers(targets v1.TargetsResult) map[string]map[string]struct{} {
	scrapeConfigsIdentifiers := make(map[string]map[string]struct{})
	for _, target := range targets.Active {
		// Generate target identifier
		identifier := generateIdentifier(string(target.Labels[Namespace]), string(target.Labels[Job]))
		splitedScrapePool := strings.Split(target.ScrapePool, "/")
		scrapeConfig := splitedScrapePool[1] + "/" + splitedScrapePool[2]
		if _, mapInitialized := scrapeConfigsIdentifiers[scrapeConfig]; !mapInitialized {
			scrapeConfigsIdentifiers[scrapeConfig] = make(map[string]struct{})
		}
		identifiers := scrapeConfigsIdentifiers[scrapeConfig]
		if _, exists := identifiers[identifier]; !exists {
			identifiers[identifier] = struct{}{}
		}
	}

	return scrapeConfigsIdentifiers
}

func ScrapeConfigsMetrics(scrapeConfigIdentifiers, metricsIdentifiers map[string]map[string]struct{}) map[string][]string {
	scrapeConfigMetrics := make(map[string][]string)
	for scrapeConfig, scIdentifiers := range scrapeConfigIdentifiers {
		var metricsExposed []string
		for scIdentifier := range scIdentifiers {
			for metric, mIdentifier := range metricsIdentifiers {
				if _, ok := mIdentifier[scIdentifier]; ok {
					metricsExposed = append(metricsExposed, metric)
				}
			}
		}
		scrapeConfigMetrics[scrapeConfig] = metricsExposed
	}
	return scrapeConfigMetrics
}
