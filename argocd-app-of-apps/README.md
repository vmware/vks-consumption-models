# VKS GitOps — App of Apps on vSphere Kubernetes Service

GitOps-driven lifecycle management for VKS clusters, add-ons, and workloads using Argo CD on the vSphere Supervisor.

This repository implements the **Argo CD App of Apps pattern** to declaratively manage the full stack — from vSphere Namespace registration, through VKS cluster provisioning with standard package add-ons, to multi-application deployment — all from a single Git repository.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Git Repository                              │
│                                                                      │
│  root-app.yaml ──► bootstrap/                                        │
│                      ├── vSphereNS/            (sync-wave: 0)        │
│                      │   └── workload-ns       Register namespace    │
│                      ├── argoApps/             (sync-wave: 10, 12)   │
│                      │   ├── cl01-app          Provision cl01        │
│                      │   └── cl02-app          Provision cl02        │
│                      └── workloads/            (sync-wave: 20, 25)   │
│                          ├── app-1             nginx → cl01          │
│                          └── app-2             online-boutique → cl01│
│                                                                      │
│  clusters/                                                           │
│    ├── cl01/                                                         │
│    │   ├── cl01.yaml              Cluster (2 workers)                │
│    │   ├── argo-register.yaml     Auto-register with Argo CD         │
│    │   ├── cert-manager-addon.yaml  VKS AddonInstall                 │
│    │   └── contour-addon.yaml       VKS AddonInstall                 │
│    └── cl02/                                                         │
│        ├── cl02.yaml              Cluster (autoscaler: 1-3)          │
│        └── argo-register.yaml     Auto-register with Argo CD         │
│                                                                      │
│  workloads/cl01/                                                     │
│    ├── nginx/                     Simple nginx deployment            │
│    └── online-boutique/           Google microservices demo (11 svc) │
└──────────────────────────────────────────────────────────────────────┘
```

## How It Works

A single `root-app.yaml` is applied to the Argo CD instance running on the Supervisor. Argo CD scans the `bootstrap/` directory and discovers child Applications, each with a sync wave that controls ordering:

| Wave | Resource | What Happens |
|------|----------|-------------|
| **0** | `ArgoNamespace` CR | Registers `workload-ns` with the Argo CD instance via the auto-attach service |
| **10** | Application `cl01-app` | Syncs `clusters/cl01/` — provisions VKS cluster, registers with Argo CD, installs cert-manager and contour |
| **12** | Application `cl02-app` | Syncs `clusters/cl02/` — provisions VKS cluster with cluster autoscaler enabled, registers with Argo CD |
| **20** | Application `app-1` | Deploys nginx into the `demo-app` namespace on cl01 |
| **25** | Application `app-2` | Deploys Google Online Boutique (11 microservices) into the `online-boutique` namespace on cl01 |

All Application and ArgoNamespace resources include the `resources-finalizer.argocd.argoproj.io` finalizer, enabling **foreground cascading deletion** — when a parent app is deleted, all child resources are cleaned up automatically.

## Clusters

### cl01 — Standard Cluster with Add-ons and Workloads

| Parameter | Value |
|-----------|-------|
| ClusterClass | `builtin-generic-v3.6.0` |
| Kubernetes Version | `v1.35.0---vmware.2-vkr.4` |
| Control Plane | 1 replica |
| Worker Pool | 2 nodes (`node-pool-1`) |
| VM Class | `best-effort-medium` |
| Storage Class | `vsan-esa-default-policy-raid5` |
| CNI | Antrea |
| Add-ons | cert-manager, contour |
| Autoscaler | Disabled |
| Workloads | nginx (app-1), Online Boutique (app-2) |

### cl02 — Autoscaler-Enabled Cluster

| Parameter | Value |
|-----------|-------|
| ClusterClass | `builtin-generic-v3.6.0` |
| Kubernetes Version | `v1.35.0---vmware.2-vkr.4` |
| Control Plane | 1 replica |
| Worker Pool | Autoscaled (min: 1, max: 3) |
| VM Class | `best-effort-medium` |
| Storage Class | `vsan-esa-default-policy-raid5` |
| CNI | Antrea |
| Add-ons | None (base cluster) |
| Autoscaler | **Enabled** via MachineDeployment annotations |

## Workloads

### app-1: nginx (wave 20)

A simple nginx-unprivileged deployment with a LoadBalancer service. Deployed to the `demo-app` namespace on cl01. Uses the Broadcom-hosted container image (`dockerhub.packages.vcfd.broadcom.net`). Includes production security context: non-root, no privilege escalation, all capabilities dropped, seccomp RuntimeDefault.

### app-2: Google Online Boutique (wave 25)

A full microservices e-commerce demo application with 11 services deployed to the `online-boutique` namespace on cl01:

| Service | Port | Description |
|---------|------|-------------|
| frontend | 8080 | Web UI (LoadBalancer exposed via `frontend-external`) |
| cartservice | 7070 | Shopping cart (backed by redis-cart) |
| productcatalogservice | 3550 | Product listing |
| currencyservice | 7000 | Currency conversion |
| paymentservice | 50051 | Payment processing |
| shippingservice | 50051 | Shipping cost calculation |
| emailservice | 8080 | Order confirmation emails |
| checkoutservice | 5050 | Checkout orchestration |
| recommendationservice | 8080 | Product recommendations |
| adservice | 9555 | Ad serving |
| redis-cart | 6379 | Cart data store |
| loadgenerator | — | Synthetic traffic generator |

All services use security-hardened containers (non-root, read-only root filesystem, capabilities dropped). The loadgenerator waits for the frontend to be reachable before starting traffic.

## VKS Add-on Management

Standard packages are installed declaratively via `AddonInstall` resources applied to the **Supervisor**. VKS automatically selects the latest compatible version and handles lifecycle management during cluster upgrades.

### Installed Add-ons

| Cluster | Add-on | AddonInstall | Installs Into |
|---------|--------|-------------|---------------|
| cl01 | cert-manager | `cl01-cert-manager` | `cert-manager` namespace |
| cl01 | contour | `cl01-contour` | `projectcontour` namespace |

### Available Add-ons

| Add-on | `addonRef.name` | Purpose |
|--------|-----------------|---------|
| cert-manager | `cert-manager` | TLS certificate automation |
| contour | `contour` | Ingress controller (Envoy-based) |
| prometheus | `prometheus` | Metrics and alerting |
| cluster-autoscaler | `cluster-autoscaler` | Automatic node scaling |
| external-dns | `external-dns` | DNS record automation |
| ako | `ako` | NSX ALB integration |
| telegraf | `telegraf` | Node metrics collection |

## Cascading Deletion

All Argo CD Application and ArgoNamespace resources include the finalizer:

```yaml
finalizers:
  - resources-finalizer.argocd.argoproj.io
```

This ensures **foreground cascading deletion**: deleting the root app triggers deletion of all child apps, which triggers deletion of all managed resources (clusters, add-ons, workloads). Removing a cluster directory from Git and syncing will tear down the cluster and all its workloads cleanly.

## Prerequisites

- **VCF 9.0+** with Supervisor enabled
- **VKS 3.5.0+** (for add-on management; ClusterClass `builtin-generic-v3.5.0` or later)
- **Argo CD Supervisor Service** — [VMware Argo CD Operator](https://vsphere-tmm.github.io/Supervisor-Services/#argocd-operator)
- **Argo CD Auto-Attach Service** — [warroyo/argocd-attach-service](https://github.com/warroyo/argocd-attach-service)
- **Broadcom customized Argo CD CLI** (version must contain `vcf` suffix)
- A vSphere Namespace for the Argo CD instance (e.g., `argocd-instance-1`)
- Git repository accessible from the Supervisor network

## Quick Start

### 1. Verify Argo CD is Running

```bash
kubectl get pods -n argocd-instance-1
```

### 2. Log in to Argo CD

```bash
# Get the LoadBalancer IP
kubectl get svc argocd-server -n argocd-instance-1

# Get the admin password
kubectl get secret -n argocd-instance-1 argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d

# Log in (use the customized Broadcom CLI)
argocd login <ARGOCD-LB-IP>
```

### 3. Register the Argo CD Instance's Own Namespace

```bash
argocd cluster add <SUPERVISOR-IP-OR-HOSTNAME> \
  --namespace argocd-instance-1 \
  --kubeconfig sc.kubeconfig
```

### 4. Apply the Root Application

This is the **only manual step**. Everything else is GitOps.

```bash
kubectl apply -f root-app.yaml
```

### 5. Watch the Cascade

```bash
# Watch all apps
argocd app list

# Watch cluster provisioning
kubectl get clusters -n workload-ns -w

# Verify add-ons
kubectl get clusteraddon -n workload-ns

# Watch workloads
argocd app get app-1
argocd app get app-2
```

### 6. Access Online Boutique

Once app-2 is healthy, get the frontend LoadBalancer IP:

```bash
argocd app get app-2
# Or from the VKS cluster directly:
kubectl get svc frontend-external -n online-boutique --kubeconfig <cl01-kubeconfig>
```

Open `http://<EXTERNAL-IP>` in your browser to access the Online Boutique storefront.

## Repository Structure

```
.
├── root-app.yaml                              # Entry point — apply once
├── bootstrap/                                 # Child Application definitions
│   ├── vSphereNS/
│   │   └── workload-ns.yaml                   # ArgoNamespace CR (wave 0)
│   ├── argoApps/
│   │   ├── cl01-app.yaml                      # App → clusters/cl01 (wave 10)
│   │   └── cl02-app.yaml                      # App → clusters/cl02 (wave 12)
│   └── workloads/
│       ├── app-1.yaml                         # App → workloads/cl01/nginx (wave 20)
│       └── app-2.yaml                         # App → workloads/cl01/online-boutique (wave 25)
├── clusters/                                  # VKS cluster definitions
│   ├── cl01/
│   │   ├── cl01.yaml                          # Cluster API v1beta2 (2 workers)
│   │   ├── argo-register.yaml                 # ArgoCluster CR (auto-attach)
│   │   ├── cert-manager-addon.yaml            # AddonInstall: cert-manager
│   │   └── contour-addon.yaml                 # AddonInstall: contour
│   └── cl02/
│       ├── cl02.yaml                          # Cluster with autoscaler (1-3 nodes)
│       └── argo-register.yaml                 # ArgoCluster CR (auto-attach)
└── workloads/                                 # Application workloads per cluster
    └── cl01/
        ├── nginx/
        │   ├── namespace.yaml                 # demo-app namespace
        │   ├── deployment.yaml                # nginx-unprivileged (3 replicas)
        │   └── service.yaml                   # LoadBalancer service
        └── online-boutique/
            └── online-boutique.yaml           # Full microservices demo (11 services)
```

## Sync Wave Ordering

```
Wave  0:  Register vSphere Namespace with Argo CD
Wave 10:  Provision cl01 + cert-manager + contour + register with Argo CD
Wave 12:  Provision cl02 (autoscaler) + register with Argo CD
Wave 20:  Deploy nginx into cl01
Wave 25:  Deploy Online Boutique into cl01
```

## Day 2 Operations

### Scale worker nodes

```yaml
# clusters/cl01/cl01.yaml
replicas: 4  # was 2
```

### Upgrade Kubernetes version

```yaml
# clusters/cl01/cl01.yaml
version: v1.36.0---vmware.x-vkr.x
```

### Add a standard package

Create `clusters/cl01/prometheus-addon.yaml`:

```yaml
apiVersion: addons.kubernetes.vmware.com/v1alpha1
kind: AddonInstall
metadata:
  name: cl01-prometheus
  namespace: workload-ns
spec:
  addonRef:
    name: prometheus
  clusters:
  - selector:
      matchLabels:
        cluster.x-k8s.io/cluster-name: cl01
```

### Add a new cluster

1. Create `clusters/cl03/` with cluster YAML and argo-register
2. Create `bootstrap/argoApps/cl03-app.yaml`
3. Push — root app discovers and provisions automatically

### Delete a cluster

Remove the cluster directory and its bootstrap app. Push. Cascading deletion tears down everything.

### Add a new workload to an existing cluster

1. Add manifests under `workloads/cl01/<new-app>/`
2. Create `bootstrap/workloads/app-N.yaml` with appropriate sync wave
3. Push — root app discovers the new child app
