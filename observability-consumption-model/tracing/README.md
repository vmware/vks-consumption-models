## Tracing ##


# OpenSearch

For the opensearch Helm values, see 'opensearch-values.yaml'
```
# Create ns for OpenSearch
NS=tracing
kubectl create ns $NS
kubectl label --overwrite ns $NS \
  pod-security.kubernetes.io/enforce=baseline

# Add OpenSearch charts repo
helm repo add opensearch https://opensearch-project.github.io/helm-charts 

# Install/upgrade Opensearch
helm upgrade --install opensearch opensearch/opensearch \
  --namespace tracing \
  --version 3.5.0 \
  -f opensearch-values.yaml \
  --wait --timeout 20m

```

# Jaeger

```
# Create ns for OpenTel Operator
NS=opentelemetry-operator-system
kubectl create ns $NS
kubectl label --overwrite ns $NS \
  pod-security.kubernetes.io/enforce=baseline

# Install the OpenTel Operator
kubectl apply -f \
  https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/\
  opentelemetry-operator.yaml

# Install OTel Operator with Jaeger v2
kubectl apply -f - <<EOF
apiVersion: opentelemetry.io/v1beta1
kind: OpenTelemetryCollector
metadata:
  name: jaeger
  namespace: tracing
spec:
  # Run Jaeger v2 as a single in-cluster deployment
  mode: deployment
  replicas: 1
  image: jaegertracing/jaeger:2.16.0

  # Expose the Jaeger UI and OTLP ingest endpoints
  ports:
    - name: jaeger-ui
      port: 16686
    - name: otlp-grpc
      port: 4317
    - name: otlp-http
      port: 4318

  config:
    service:
      # Enable storage, query/UI, and health endpoints
      extensions: [jaeger_storage, jaeger_query, healthcheckv2]

      pipelines:
        traces:
          # Receive OTLP traces and persist them via the Jaeger storage exporter
          receivers: [otlp]
          processors: [batch]
          exporters: [jaeger_storage_exporter]

      telemetry:
        resource:
          service.name: jaeger
        metrics:
          # Expose basic Prometheus metrics for the collector itself
          level: basic
          readers:
            - pull:
                exporter:
                  prometheus:
                    host: 0.0.0.0
                    port: 8888
        logs:
          level: info

    extensions:
      # Health endpoint used for readiness / liveness checks
      healthcheckv2:
        use_v2: true
        http: {}

      # Jaeger query service and UI
      jaeger_query:
        http:
          endpoint: 0.0.0.0:16686
        storage:
          traces: opensearch_store

      # OpenSearch backend used for persistent trace storage
      jaeger_storage:
        backends:
          opensearch_store:
            opensearch:
              server_urls:
                - https://opensearch-cluster-master.tracing.svc.cluster.local:9200
              tls:
                # Simplified for this implementation -- 
                # replace with full cert validation in production
                insecure_skip_verify: true
              auth:
                basic:
                  username: admin
                  password: "C0rrugated!Trace#2026"
              indices:
                index_prefix: jaeger
                spans:
                  shards: 1
                  replicas: 0
                services:
                  shards: 1
                  replicas: 0
                dependencies:
                  shards: 1
                  replicas: 0
                sampling:
                  shards: 1
                  replicas: 0

    receivers:
      otlp:
        # Accept OTLP traces over both gRPC and HTTP
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318

    processors:
      # Batch trace data before writing to storage
      batch: {}

    exporters:
      # Persist traces into the configured Jaeger storage backend
      jaeger_storage_exporter:
        trace_storage: opensearch_store
EOF


# Label the monitoring namespace so the Gateway trusts it
kubectl label ns tracing access=gateway-trusted --overwrite

# Apply the gateway and routes
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: jaeger-gateway
  namespace: istio-ingress
spec:
  gatewayClassName: istio
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      hostname: jaeger.lab
      tls:
        mode: Terminate
        certificateRefs:
          - name: monitoring-gateway-tls
      allowedRoutes:
        namespaces:
          from: Selector
          selector:
            matchLabels:
              access: gateway-trusted
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: jaeger
  namespace: tracing
spec:
  parentRefs:
    - name: jaeger-gateway
      namespace: istio-ingress
  hostnames:
    - jaeger.lab
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: jaeger-collector
          port: 16686
EOF  
```

