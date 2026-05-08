# Observability on vSphere Kubernetes Service

This repository contains the example manifests, Helm values, and supporting configuration used in the **Observability on vSphere Kubernetes Service** white paper.

**White paper:**  
https://www.vmware.com/docs/observability-vks

## Overview

This repository demonstrates a practical observability architecture for **vSphere Kubernetes Service (VKS)** on **VMware Cloud Foundation (VCF)**, built around the three pillars of observability:

- **Metrics** using Prometheus, Grafana, Istio telemetry, and VCF Operations
- **Logging** using Fluent Bit, Grafana Loki, Grafana, and VCF Operations for Logs
- **Tracing** using OpenTelemetry, Jaeger v2, and OpenSearch

The content is organized to mirror the structure of the white paper and provide reusable implementation examples for VKS environments.

## Repository Structure

### `metrics/`

Artifacts related to the metrics pipeline and validation workflow, including:

- Prometheus Community Stack (`kube-prometheus-stack`)
- Gateway API configuration for Grafana and Prometheus
- ServiceMonitor and PodMonitor examples
- Sample application manifests for metrics scraping

### `logging/`

Artifacts related to Kubernetes log collection and forwarding, including:

- Loki Helm values
- Fluent Bit configuration
- Dual-destination log forwarding to Loki and VCF Operations for Logs

### `tracing/`

Artifacts related to distributed tracing and trace storage, including:

- OpenSearch configuration
- Jaeger v2 deployment via the OpenTelemetry Operator
- Gateway API configuration for Jaeger
- OpenTelemetry demo-related tracing artifacts

## Getting Started

1. Review the white paper for the overall architecture and workflow.
2. Navigate to the relevant topic directory:
   - `metrics/`
   - `logging/`
   - `tracing/`
3. Apply or adapt the manifests for your own VKS environment.
4. Use the paper appendices for supporting setup and validation details.

## Notes

These examples are intended to accompany the reference implementation described in the paper. Some values may need to be adapted for your environment, including:

- namespaces
- hostnames and DNS
- storage classes
- certificate issuers
- registry locations
- infrastructure endpoints

## References
* [vSphere Supervisor Platform](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-9-0-and-later/9-0/vsphere-supervisor-installation-and-configuration.html)
* [Command line tool (kubectl)](https://kubernetes.io/docs/reference/kubectl/)
* [Installing and Using VCF CLI v9.0](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-9-0-and-later/9-0/building-your-cloud-applications/getting-started-with-the-tools-for-building-applications/installing-and-using-vcf-cli-v9.html)
* [Managing Add-ons in VKS](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/managing-add-ons-in-vks-clusters.html)
* [VKS Standard Package Reference](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/installing-standard-packages-on-tkg-service-clusters/standard-package-reference.html)

## Purpose

The goal of this repository is to make the configuration artifacts from the paper easier to review, reuse, and adapt, while keeping the structure aligned to the core observability domains of **metrics**, **logging**, and **tracing**.