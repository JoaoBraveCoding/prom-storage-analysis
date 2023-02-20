# prom-storage-analysis


## How to run

### Metrics Per ServiceMonitor 

Print a list of metrics being used on rules and alerts per ServiceMonitor.

```bash
# You have Prometheus on localhost:9090
go run main.go
# If you have Prometheus running in another machine
go run main.go --url https://prom.example.org:9090
# If your Prometheus is behind https you can pass a bearer token
go run main.go --url https://prom.example.org:9090 --bearer-token ZUYSEWdasDdawdFEAEF88d7DWQ9dad8a
# If you want to pass it a start and ending time to only show series from that time window
go run main.go --start 2022-07-29T14:00:00.00Z --end 2022-07-29T15:15:00.00Z
# If you want to use an Promecieus endpoint you should the path to a rules-file from the mustgather
go run main.go --url https://rxnxxkzs-promecieus.apps.cr.j7t7.p1.openshiftapps.com --rules-file mustgather/quay.../rules.json
```

### Account for dashboards

Some of the metrics used by the OpenShift console live in a dashboard file,
these dashboards are in format that it's understood by Grafana. However, we are
only interested in the expressions present in these dashboards, but before
passing that expression to the prometheus parser we have to replace grafana
variables (variables that start with $) that are used to define time intervals or
rate intervals.
Once a file is generated we can pass it to the script using the cmd "expr-file".

```shell
cat ../cluster-monitoring-operator/manifests/0000_90_cluster-monitoring-operator_01-dashboards.yaml | grep -oP "(?<=\"expr\": \").*(?=,)" | sed -e 's/^"//' -e 's/"$//' | sed -e 's/\\"/"/g' | sed -e "s/\$interval/1h/g" | sed -e "s/\$resolution/5m/g" | sed -e "s/\$__rate_interval/5m/g" | sed -e "s/\\\n//g" > expr-list.txt
```

Then we can pass this file to the script, to print a list of metrics being used in 
alerts, rules and dashboards organized per ServiceMonitor

```shell
go run main.go --cmd=all --expressions-file=expr-list.txt > all-metrics-per-sm.txt
```