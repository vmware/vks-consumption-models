# Logging #

## Install Loki

See the 'loki-values.yaml' for the Loki Helm chart values
```
# Add Grafana charts repo
helm repo add grafana https://grafana.github.io/helm-charts

# Install/upgrade Loki
helm upgrade --install loki grafana/loki \
  -n loki --create-namespace \
  -f loki-values.yaml \
  --wait --timeout 20m
```

## Fluent-bit

Package reference
https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/installing-standard-packages-on-tkg-service-clusters/standard-package-reference/fluent-bit-package-reference.html

```
# Confirm fluent-bit package is installed
kubectl -n vmware-system-tkg describe \
  $(kubectl get pkgi -A -o name | grep fluent-bit) | grep -A5 'Status'

# Check runtime
kubectl get pods -A -l app=fluent-bit

# Create a secret with VCF/logs Certificate
kubectl -n tanzu-system-logging create secret generic tls-ca-cert \
  --from-file=tls_http.crt=vcf-ops-logs-ca.pem


# Create the Fluent Bit values file
# Outputs to both Loki & VCF Ops/Logs 
VCF_LOGS_HOST=ops-logs.lab
VKSCLUSTER=vks-kubernetes-cluster

cat << EOF > fluent-bit-values.yaml
fluent_bit:
  config:
    outputs: |
      [OUTPUT]
        Name          http
        Match         *
        Host          ${VCF_LOGS_HOST}
        Port          9543
        URI           api/v2/events
        Format        json
        Header        Content-Type application/json
        tls           on
        tls.debug     4
        tls.verify    on
        tls.ca_file   /etc/ssl/certs/tls_http.crt
        Retry_Limit   False

      [OUTPUT]
        Name          loki
        Match         kube.*
        Host          cluster-loki-gateway.loki.svc.cluster.local
        Port          80
        URI           /loki/api/v1/push
        labels        job=fluent-bit,source=vks,cluster=${VKSCLUSTER}
        label_keys    \$namespace_name,\$pod_name,\$container_name,\$host,\$stream

  daemonset:
    secretName: tls-ca-cert

namespace: tanzu-system-logging
EOF

# Switch to the supervisor context
vcf context use <supervisor namespece>

# Update the fluent-bit config
vcf addon install update fluent-bit \
  --cluster-name $VKSCLUSTER \
  -f fluent-bit-values.yaml

# Switch back to the VKS context
vcf context use <vks cluster>

# Restart Fluent-bit daemonset to reconcile changes
kubectl -n tanzu-system-logging rollout restart daemonset fluent-bit

# Ensure pods are healthy
kubectl -n tanzu-system-logging get pods
```