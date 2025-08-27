# Observability Setup for GAuth

This directory contains configuration files for the observability stack used by GAuth. The stack includes:

- **OpenTelemetry Collector**: Collects traces, metrics, and logs from GAuth and forwards them to the appropriate backends.
- **Prometheus**: Stores metrics data.
- **Loki**: Stores logs data.
- **Promtail**: Collects logs from Docker containers.
- **Jaeger**: Stores and visualizes distributed traces.
- **Grafana**: Visualizes metrics, logs, and traces.

## Components

### OpenTelemetry Collector

The OpenTelemetry Collector is a vendor-agnostic way to collect, process, and export telemetry data. It receives data from GAuth and forwards it to the appropriate backends.

Configuration: `otel-collector-config.yaml`

### Prometheus

Prometheus is a time series database for storing metrics data. It scrapes metrics from GAuth and the OpenTelemetry Collector.

Configuration: `prometheus.yml`

### Loki

Loki is a log aggregation system designed for storing and querying logs.

Configuration: `loki-config.yaml`

### Promtail

Promtail is an agent that ships the contents of local logs to Loki.

Configuration: `promtail-config.yaml`

### Jaeger

Jaeger is a distributed tracing system for monitoring and troubleshooting microservices-based distributed systems.

### Grafana

Grafana is a visualization tool that allows you to query, visualize, alert on, and understand your metrics, logs, and traces.

Configuration:
- `grafana-datasources.yaml`: Configures data sources for Grafana.
- `grafana-dashboards.yaml`: Configures dashboard provisioning.
- `dashboards/`: Contains pre-configured dashboards.

## Usage

The observability stack is configured in the main `docker-compose.yml` file. To start the stack, run:

```bash
docker-compose up -d
```

### Accessing the UIs

- **Grafana**: http://localhost:3000 (username: admin, password: admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

## Dashboards

The following dashboards are available in Grafana:

- **GAuth Overview**: Shows key metrics for the GAuth service, including request rates, latencies, and business metrics.

## Metrics

GAuth exposes the following metrics:

- **HTTP metrics**: Request counts, latencies, and status codes.
- **gRPC metrics**: Request counts, latencies, and status codes.
- **Database metrics**: Query counts, latencies, and error rates.
- **Renclave client metrics**: Request counts, latencies, and error rates.
- **Business metrics**: Counts of wallets, private keys, organizations, users, and activities created.

## Traces

GAuth uses OpenTelemetry to instrument the following components:

- **HTTP server**: All HTTP requests.
- **gRPC server**: All gRPC requests.
- **Renclave client**: All requests to the Renclave service.

## Logs

Logs from all services are collected by Promtail and stored in Loki. The logs are structured in JSON format and include trace IDs for correlation with traces.

