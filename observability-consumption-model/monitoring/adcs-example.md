## Cert-manager with ADCS Integration

```yaml
### 1. Prerequisites: Namespace & Credentials
kubectl create ns adcs-issuer
kubectl label --overwrite ns adcs-issuer pod-security.kubernetes.io/enforce=baseline

# Create secret containing user/pass for windows CA
# Ensure account has the correct rights to 'enroll' certificates (see below)
kubectl -n adcs-issuer create secret generic adcs-issuer-credentials \
  --from-literal=username='domain\[your username]' \
  --from-literal=password='[your password]'

### 2. Install the ADCS Issuer Webhook
helm repo add djkormo-adcs-issuer https://djkormo.github.io/adcs-issuer/
helm repo update

# Inspect available versions
helm search repo djkormo-adcs-issuer/adcs-issuer --versions

# Pull default values for a version you want to pin
helm show values djkormo-adcs-issuer/adcs-issuer --version 2.2.1 > adcs-values.yaml

# Install (default values in adcs-values.yaml will server most environments)
helm upgrade --install adcs-issuer djkormo-adcs-issuer/adcs-issuer \
  --namespace adcs-issuer \
  --create-namespace \
  --version 2.2.1 \
  --values adcs-values.yaml

### 3. Configure the Cluster Issuer
# Create clusteradcsissuer: first obtain a base64 encoded certificate 
# From the Windows CA & encode with 'base64 -w0 certificate.cer'
# Then add to the key 'caBundle' 
# NB: ensure account used has permissions to use the "webserver" template
kubectl apply -f - <<EOF
apiVersion: adcs.certmanager.csf.nokia.com/v1
kind: ClusterAdcsIssuer
metadata:
  name: adcs-cluster-issuer
spec:
  caBundle: 
  credentialsRef:
    name: adcs-issuer-credentials
  statusCheckInterval: 5m
  retryInterval: 5m
  url: "https://lvn-sc-ca01.showcase.tmm.broadcom.lab/certsrv"
  templateName: "WebServer"
EOF

### 4. Request a Certificate
# Here we request a certificate for an istio ingress gateway
# and store as the secret 'monitoring-gateway-tls'
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

### 5. Consume the Certificate via Gateway API
# Apply the above to a gateway listener under the 'tls/certifcateRefs' key
# Only allow namespaces with the label ‘gateway-trusted’
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
