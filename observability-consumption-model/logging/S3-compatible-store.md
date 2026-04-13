## S3 Compatible Store (MinIO)

```
# Add MinIO Helm chart
# Note: MinIO charts (charts.min.io) are being 
# replaced by helm.min.io which use Alstor
helm repo add minio https://charts.min.io/

# Create namespace & set permissions
NS=minio
kubectl create ns $NS
kubectl label --overwrite ns $NS pod-security.kubernetes.io/enforce=baseline

# Create secrets for MinIO root user and Loki user
miniorootpw=<pass for minio root>
lokiuserpw=<pass for loki user>

kubectl apply -f - << EOF
apiVersion: v1
kind: Secret
metadata:
  name: minio-root-credentials
  namespace: minio
type: Opaque
stringData:
  rootUser: minioadmin
  rootPassword: "$miniorootpw"
---
apiVersion: v1
kind: Secret
metadata:
  name: minio-loki-credentials
  namespace: minio
type: Opaque
stringData:
  password: "$lokiuserpw"
EOF

# create MinIO manifest
cat << EOF > custom-minio-values-5.4.0.yaml
mode: distributed          
replicas: 2                
pools: 1                  
drivesPerNode: 2 

existingSecret: minio-root-credentials

persistence:
  enabled: true
  storageClass: vsan-esa-default-policy-raid5
  accessMode: ReadWriteOnce
  size: 100Gi   # per pod. With 4 pods this is 400Gi raw.

# limit resources
resources:
  requests:
    cpu: 100m
    memory: 512Mi
  limits:
    cpu: "1"
    memory: 2Gi

# Buckets Loki commonly expects (chunks/ruler/admin).
buckets:
  - name: loki-chunks
    policy: none
    purge: false
    versioning: false
  - name: loki-ruler
    policy: none
    purge: false
    versioning: false
  - name: loki-admin
    policy: none
    purge: false
    versioning: false

# Loki access key
users:
  - accessKey: loki
    existingSecret: minio-loki-credentials
    existingSecretKey: password   
    policy: readwrite   # built-in policy    
EOF

# Install using Helm
helm install minio minio/minio \
  -n minio \
  --version 5.4.0 \
  -f custom-minio-values-5.4.0.yaml
```