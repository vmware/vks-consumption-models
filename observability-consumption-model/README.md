# Observability on vSphere Kubernetes Service

This repository provides the example manifests, Helm values, and supporting configuration used in the **Observability on vSphere Kubernetes Service** white paper.

For the related white paper, visit:
https://www.vmware.com/docs/observability-vks

* * *

## Overview

The repository demonstrates a practical observability architecture for **vSphere Kubernetes Service (VKS)** on **VMware Cloud Foundation (VCF)**, built around the three pillars of observability:

1. **Metrics** using Prometheus, Grafana, Istio service telemetry, and VCF Operations
2. **Logging** using Fluent Bit, Grafana Loki, Grafana, and VCF Operations for Logs
3. **Tracing** using OpenTelemetry, Jaeger v2, and OpenSearch

The goal is to provide reusable implementation examples that can be reviewed, adapted, and deployed in VKS environments.

* * *

## Repository Structure

The repository is organized into three functional areas:

### 1. `metrics/`

Contains the manifests and values used to deploy and validate the metrics stack.

Examples include:
- Prometheus Community Stack (`kube-prometheus-stack`)
- Gateway API configuration for Grafana and Prometheus
- ServiceMonitor and PodMonitor examples
- Supporting sample application manifests

### 2. `logging/`

Contains the configuration used to deploy and integrate the logging stack.

Examples include:
- Loki Helm values
- Fluent Bit configuration
- Dual-destination log forwarding to Loki and VCF Operations for Logs
- Supporting integration artifacts

### 3. `tracing/`

Contains the manifests and values used to deploy the tracing stack.

Examples include:
- OpenSearch configuration
- Jaeger v2 deployment via the OpenTelemetry Operator
- Gateway API configuration for Jaeger
- OpenTelemetry demo-related tracing artifacts

* * *

## Getting Started

1. Read the white paper to understand the overall architecture and workflow.
2. Navigate to the relevant topic directory:
   - `metrics/`
   - `logging/`
   - `tracing/`
3. Apply or adapt the manifests for your own VKS environment.
4. Refer to the paper appendices for supporting setup details and validation steps.

* * *

## Notes

These examples are intended to accompany the reference implementation described in the paper. Some values may need to be adjusted for your environment, including:

- namespaces
- hostnames and DNS
- storage classes
- certificate issuers
- registry locations
- infrastructure endpoint details

* * *

## Key Benefits

- **Topic-based structure** aligned to metrics, logs, and traces
- **Reusable manifests and values** that map directly to the white paper
- **Kubernetes-native approach** built around familiar open source tooling
- **Infrastructure-aware observability** through integration with VCF Operations and VCF Operations for Logs