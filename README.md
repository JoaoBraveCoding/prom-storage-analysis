# prom-storage-analysis


## How to run

```bash
# If you have Prometheus on localhost:9090
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