# prom-storage-analysis


## How to run

```bash
# If you have Prometheus on localhost:9090
go run cmd/main.go
# If you have Prometheus running in another machine
go run cmd/main.go https://prom.example.org:9090
```