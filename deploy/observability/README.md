# Observability Stack

This directory contains configuration for the observability stack:
- **Loki**: Log aggregation (stores JSON logs)
- **Tempo**: Distributed tracing (stores spans for tree-view)
- **Grafana**: Visualization (Explore logs & traces)

## Quick Start

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Loki | http://localhost:3100 | - |
| Tempo | http://localhost:3200 | - |

## Two Modes of Tracing

### Mode 1: Log-based Tracing (Simple)

Just use JSON logs with trace_id. Logs appear as a **list** in Grafana.

```go
span, ctx := tracer.ServiceSpan(ctx, log, "Register")
defer span.End()
```

**Query in Grafana:**
```logql
{service="classic-api"} |= `trace_id="abc123"`
```

### Mode 2: OTEL-based Tracing (Tree View)

Use OpenTelemetry spans. Traces appear as a **tree** in Grafana Tempo.

```go
// Initialize at startup
shutdown, _ := tracer.InitOTEL(ctx, tracer.Config{
    ServiceName:  "classic-api",
    OTLPEndpoint: "localhost:4317",
    SampleRate:   1.0, // 100% sampling for dev
})
defer shutdown(ctx)

// Use unified span (logs + OTEL)
span, ctx := tracer.UnifiedServiceSpan(ctx, log, "Register")
defer span.End()
```

**View in Grafana:**
1. Go to **Explore** > Select **Tempo** datasource
2. Search by trace_id or service name
3. Click trace to see **tree-view** with timing

## Architecture

```
Application
    |
    +-- zerolog --> JSON logs --> Promtail --> Loki
    |                                            |
    |                                            v
    +-- OTEL SDK --> OTLP --> Tempo --> Grafana
```

## Grafana Features

### 1. Explore Logs (Loki)

```logql
# All errors
{service="classic-api"} | json | level = "error"

# By trace_id
{service="classic-api"} |= `trace_id="..."`

# By operation
{service="classic-api"} | json | operation = "service:Register"

# Slow requests (>100ms latency)
{service="classic-api"} | json | latency >= "100ms"
```

### 2. Explore Traces (Tempo)

- Search by TraceID
- Search by Service
- Search by Span name
- View waterfall/timeline

### 3. Derived Fields (Link Logs to Traces)

Click the "View Trace" button next to any log line to jump to the tree-view trace.

## Configuration

### Application Config

Add to your config.yaml:

```yaml
trace:
  enabled: true
  otlp_endpoint: "localhost:4317"
  sample_rate: 1.0

log:
  level: info
  output: ./logs  # For Promtail to scrape
```

### Environment Variables

```bash
# Enable OTEL tracing
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# Log output directory
export LOG_DIR=./logs
```

## Dashboard

Import the pre-built dashboard:
1. Go to Dashboards > Import
2. Upload `dashboards/classic-api-overview.json`

Panels:
- Request Logs (with trace_id)
- Log Levels Over Time
- Operations Over Time
- Error Logs

## Alerting

Example alert rules in Grafana:

```yaml
# High error rate
expr: sum(rate({service="classic-api"} | json | level = "error" [5m])) > 10
for: 5m
annotations:
  summary: "High error rate detected"
```

## Troubleshooting

### No logs appearing in Loki

1. Check Promtail is running: `docker-compose ps promtail`
2. Check logs directory is mounted: `ls -la logs/`
3. Check Promtail logs: `docker-compose logs promtail`

### No traces appearing in Tempo

1. Check OTEL is initialized in app
2. Check OTLP endpoint is correct
3. Check Tempo is receiving: `curl http://localhost:3200/ready`

### Cannot link logs to traces

1. Ensure `trace_id` field is present in logs
2. Check derived fields config in datasources
3. Verify trace_id matches between logs and spans
