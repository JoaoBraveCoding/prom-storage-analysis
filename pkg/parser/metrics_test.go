package parser

import (
	"sort"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	for _, tc := range []struct {
		name            string
		expression      string
		expectedMetrics []string
	}{
		{
			name:            "simple expression",
			expression:      "alertmanager_config_last_reload_successful",
			expectedMetrics: []string{"alertmanager_config_last_reload_successful"},
		},
		{
			name:            "simple expression with labels",
			expression:      `alertmanager_config_last_reload_successful{job=~"alertmanager-main|alertmanager-user-workload"}`,
			expectedMetrics: []string{"alertmanager_config_last_reload_successful"},
		},
		{
			name:            "simple expression with labels and comments",
			expression:      `# Without max_over_time, failed scrapes could create false negatives, see
# https://www.robustperception.io/alerting-on-gauges-in-prometheus-2-0 for details.
max_over_time(alertmanager_config_last_reload_successful{job=~"alertmanager-main|alertmanager-user-workload"}[5m]) == 0
`,
			expectedMetrics: []string{"alertmanager_config_last_reload_successful"},
		},
		{
			name:            "two expressions with labels",
			expression:      `(
rate(alertmanager_notifications_failed_total{job=~"alertmanager-main|alertmanager-user-workload"}[5m])
/
rate(alertmanager_notifications_total{job=~"alertmanager-main|alertmanager-user-workload"}[5m])
)
> 0.01
`,
			expectedMetrics: []string{"alertmanager_notifications_total", "alertmanager_notifications_failed_total"},
		},
		{
			name:            "complex expressions with labels",
			expression:      `(((
				kube_deployment_spec_replicas{namespace=~"(openshift-.*|kube-.*|default)",job="kube-state-metrics"}
				  >
				kube_deployment_status_replicas_available{namespace=~"(openshift-.*|kube-.*|default)",job="kube-state-metrics"}
			  ) and (
				changes(kube_deployment_status_replicas_updated{namespace=~"(openshift-.*|kube-.*|default)",job="kube-state-metrics"}[5m])
				  ==
				0
			  )) * on() group_left cluster:control_plane:all_nodes_ready) > 0
`,
			expectedMetrics: []string{"cluster:control_plane:all_nodes_ready", "kube_deployment_spec_replicas", "kube_deployment_status_replicas_available", "kube_deployment_status_replicas_updated"},
		},
	} {
		t.Run(tc.name, func(t* testing.T){
			resultExpressions := Metrics(tc.expression)
			sort.Strings(resultExpressions)
			sort.Strings(tc.expectedMetrics)
			for i, expectedMetric := range tc.expectedMetrics {
				if expectedMetric != resultExpressions[i] {
					t.Fatalf("Expected mtric to be %s, but got %s", expectedMetric, resultExpressions[i])
				}
			}
		})
	}
}
