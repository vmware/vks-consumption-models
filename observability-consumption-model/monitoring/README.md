# Monitoring #

## Install kube-prometheus-stack
```
# Add the Prometheus Community repo
helm repo add prometheus-community \
  https://prometheus-community.github.io/helm-charts

# Create namespace and set permissions
kubectl create ns monitoring
kubectl label --overwrite ns monitoring \
  pod-security.kubernetes.io/enforce=privileged

# Install the operator includes (Prometheus & Grafana)
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring
```

## Optionally, Without Node Exporter
```
# Create namespace with standard 'baseline' security
kubectl create ns monitoring
kubectl label --overwrite ns monitoring \
  pod-security.kubernetes.io/enforce=baseline

# Install the stack without Node Exporter
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set nodeExporter.enabled=false
```

## Verify the Installation
```
# Show running pods
kubectl -n monitoring get pods -o custom-columns=NAME:.metadata.name,\
STATUS:.status.phase,\
READY:'.status.containerStatuses[*].ready'

# Show services 
kubectl -n monitoring get svc -o custom-columns=NAME:.metadata.name,\
TYPE:.spec.type,\
IP:.spec.clusterIP,\
PORT:'.spec.ports[*].port'

# Verify service accounts
kubectl -n monitoring get sa -l app.kubernetes.io/instance=prometheus

# Verify cluster roles
kubectl get clusterroles -l app.kubernetes.io/instance=prometheus
kubectl describe clusterrole prometheus-kube-prometheus-prometheus
```

## Istio Ingress Gateway with (ADCS) Certificate.

For an example of how to setup Active Directory Certificate Services, see 'adcs-example.md'

```
# Create ns for Istio ingress gw
NS=istio-ingress
kubectl create ns $NS
kubectl label --overwrite ns $NS \
  pod-security.kubernetes.io/enforce=baseline 

# Apply the Certificate Request
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: monitoring-gateway-cert
  namespace: istio-ingress
spec:
  secretName: monitoring-gateway-tls
  commonName: "*.lab"
  dnsNames:
    - "*.lab"
  issuerRef:
    group: adcs.certmanager.csf.nokia.com
    kind: ClusterAdcsIssuer
    name: adcs-cluster-issuer
EOF


# Label the monitoring namespace so the Gateway trusts it
kubectl label ns monitoring access=gateway-trusted --overwrite

# Apply the Gateway Listener
kubectl apply -f - <<EOF
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: monitoring-gateway
  namespace: istio-ingress
spec:
  gatewayClassName: istio
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    hostname: "*.lab"
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
EOF
```

## HTTP Routes
```
kubectl apply -f - <<EOF
# 1. Grafana Route -> Service Port 80
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grafana-route
  namespace: monitoring
spec:
  parentRefs:
  - name: monitoring-gateway
    namespace: istio-ingress
    sectionName: https 
  hostnames:
  - "grafana.lab"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: prometheus-grafana
      port: 80
---

# 2. Prometheus & Alert Manager Route -> Service Ports 9090 & 9093
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: prometheus-route
  namespace: monitoring
spec:
  parentRefs:
  - name: monitoring-gateway
    namespace: istio-ingress
    sectionName: https
  hostnames:
  - "prometheus.lab"
  rules:
  
  # Rule A: Alertmanager (Sub-path with Rewrite)
  # Traffic to ‘/alertmanager’ is rewritten to ‘/’ and sent to port 9093
  - matches:
    - path:
        type: PathPrefix
        value: /alertmanager
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /
    backendRefs:
    - name: prometheus-kube-prometheus-alertmanager
      port: 9093

  # Rule B: Prometheus (Root Path)
  # Traffic to ‘/’ is sent to port 9090
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: prometheus-kube-prometheus-prometheus
      port: 9090
EOF
```

## Verify the Gateway
```
# Check for reconciliation and IP 
kubectl -n istio-ingress get gateway monitoring-gateway

# Verify routing
kubectl -n monitoring get httproutes -o \
  jsonpath='{.items[*].status.parents}' | jq

# Get the Gateway's External IP 
GATEWAY_IP=$(kubectl get gateway -n istio-ingress monitoring-gateway \
  -o jsonpath='{.status.addresses[0].value}')

# Check Grafana access
curl -ik \
  --resolve grafana.lab:443:$GATEWAY_IP https://grafana.lab/login

# Check Prometheus access & URL Re-write
curl -ik \
  --resolve prometheus.lab:443:$GATEWAY_IP https://prometheus.lab/alertmanager 
```

## Deploy Prometheus Example app
```
# Create the application namespace & set permissions
kubectl create ns my-app
kubectl label --overwrite \
  ns my-app pod-security.kubernetes.io/enforce=baseline

# Deploy the application and service
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prom-example-app
  namespace: my-app
  labels:
    app: prom-example-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prom-example-app
  template:
    metadata:
      labels:
        app: prom-example-app
    spec:
      containers:
      - name: prom-example-app
        image: fabxc/instrumented_app
        ports:
        - name: web
          containerPort: 8080
---
kind: Service
apiVersion: v1
metadata:
  name: prom-example-app
  namespace: my-app
  labels:
    app: prom-example-app
spec:
  selector:
    app: prom-example-app
  ports:
  - name: web
    port: 8080
EOF

# Create a service monitor for the app
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prom-example-app-monitor
  namespace: prometheus-operator
  labels:
    release: prometheus
spec:
  # Explicitly select the tenant namespace
  namespaceSelector:
    matchNames:
    - my-app
  selector:
    matchLabels:
      app: prom-example-app
  endpoints:
  - port: web
    interval: 15s
EOF

# Create a temporary pod to generate load
kubectl run -i --tty load-generator \
  --rm \
  --image=registry.k8s.io/e2e-test-images/busybox:1.29-2 \
  --restart=Never \
  --namespace=my-app \
  -- /bin/sh -c 
     "while sleep 0.01; do wget -q -O- http://prom-example-app:8080; echo; done"
```


## Service Mesh Observability
```
# Ensure Istio system pods are running
kubectl -n istio-system get pods | grep istio

# Label for Istio & restart the app pods
kubectl label namespace my-app istio-injection=enabled
kubectl rollout restart deployment/prom-example-app -n my-app

# Check that the Istio sidecar is ready
NS=my-app
kubectl -n $NS get pods -o name | \
  xargs -I {} kubectl -n $NS logs {} -c istio-proxy |  \
  awk '{print $4,$5,$6,$7,$8}' | \
  grep ready

# Apply a PodMonitor
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: istio-sidecars
  namespace: prometheus-operator
  labels:
    release: prometheus
spec:
  # Select all pods with the Istio sidecar injected
  selector:
    matchLabels:
      security.istio.io/tlsMode: istio
  namespaceSelector:
    matchNames:
    - my-app2
  # Target the standard Envoy metrics port
  podMetricsEndpoints:
  - path: /stats/prometheus
    port: http-envoy-prom   # Port 15090
    interval: 15s
EOF

# Deploy a stable client pod (sleeper) to act as the traffic source
kubectl run -n my-app2 debug-sleeper \
  --image=registry.k8s.io/e2e-test-images/busybox:1.29-2 \
  --restart=Never \
  -- /bin/sh -c "sleep 3600"

# Wait for the sidecar to be fully initialized (2/2 Ready)
echo "Waiting for sidecar injection..."
kubectl wait --for=condition=Ready pod/debug-sleeper -n my-app2 --timeout=60s

# Generate traffic loop (Mesh Client -> Application)
# Press Ctrl+C to stop
kubectl exec -it -n my-app2 debug-sleeper -- \
  /bin/sh -c \
  "while true; do wget -q -O- http://prom-example-app:8080; \
  echo -n .; sleep 0.1; done"
```

## Integration with VCF Operations

For Telegraf config, see 'telegraf-data-values.yaml'
```
# Find the Telegraf package and namespace
read -r NAMESPACE PKG_NAME <<< \
  $(kubectl get pkgi -A --no-headers | awk '/telegraf/{print $1, $2}')

# Update Telegraf
vcf package installed update "$PKG_NAME" \
  --values-file telegraf-data-values.yaml \
```


## Noisy Neighbor Scenario

For the workload generator, see 'critical-app.yaml'

First, create a number of Ubuntu test VMs
```
# Clone the Ubuntu image
seq 1 10 | xargs -P0 -I {} govc vm.clone -vm ubuntu ubuntu-clone{}

# Define the run script
cat << EOF  > run_all.sh
#!/bin/bash
export input=\$1
govc find -type m -name 'ubuntu-clone*' | \
  xargs -P0 -I \
  '{}' bash -c 'ssh -o "StrictHostKeyChecking=no" ubuntu@\$(govc vm.ip {}) "\$input"'
EOF
```
