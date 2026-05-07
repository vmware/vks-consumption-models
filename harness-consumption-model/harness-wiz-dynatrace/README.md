# Harness CI/CD Pipeline with Wiz & Dynatrace
This repository contains the CI/CD configuration for a multi-service microservices application. This pipeline serves as an extension of the core Harness CI/CD worflow, integrating enterprise-grade security with Wiz via STO (Security Testing Orchestration) and reliability and Observability with Dynatrace through SRM (Service Reliability Management).

---

## Pipeline Overview

The pipeline is split into two stages that handle build-time security and run-time performance monitoring.

### Build and Push (CI with STO)

* **Codebase Prep**: Clones the repository and uses a Run step to dynamically update Dockerfiles. It replaces standard base images with a private "passthrough" repository to comply with internal network/security policies. 

> **Note:** This step if optional and not required if direct Docker image pull works in your environment.

* **Security Testing Orchestration (STO)**: Pipeline utilizes the Harness STO module. This is a prerequisitie for calling Wiz scanner and having the findings automatically aggregated and deduplicated within the vulnerabilities tab of the pipeline execution.

  * **Wiz Repo Scan**: Scans the source code for secrets and vulnerabilities pre-build.
  * **Wiz Image Scan**: Post-build scanning of 12 microservices to ensure no high-risk artifacts are pushed to the registry.

* **Security (Wiz)**: Performs a pre-build repository scan and a post-build image scan using Wiz to ensure no vulnerabilities or secrets are exposed.

* **Parallel Build**: Builds and pushes 12 microservices (e.g., adservice, checkoutservice, frontend) to Artifactory.

* **GitOps Update**: Automatically clones the Helm Charts repository, updates the image tags in values-vks-agents.yaml with the current <+pipeline.executionId>, and pushes the change back to Artifactory.

### Deploy (CD with SRM)

* **Service Reliability Management (SRM)**: This pipleine requires a subscription to the Harness SRM module. SRM enabled the integration of Dyntrace as a health source. Without SRM, Harness cannot ingrest or analyze metrics from Dynatrace endpoint to determine success.

* **Rolling Update**: Performs a Kubernetes Rolling Deployment to the prodenv environment.

* **Atomated Health Tracking**: After the kubernetes rollout, SRM initiates a "Verification phase".
  
  * It pulls real-time performance data (Latenct, Error Rates, etc.) from the Dynatrace API.
  * Pipeline automatically assesses if the new version meets the reliability SLOs (Service Level Objectives).


## Prerequisites

### Harness Infrastructure

* **Harness Delegates**: Ensure delegates are running in VKS cluster.

* **Connectors**:

   * rkgithubconnector: Access to the application source code.

   * harnessk8sconnector: Access to the target VKS cluster.

   * artifactorydockerconnector: Credentials for the private Artifactory registry.

* **Secrets Management**

  Ensure the following secrets exist in your Harness Manager:

  * wiz-client-id / wiz-client-secret

  * artifactory-username / artifactory-secret

  * rk-github-secret (Personal Access Token with repo push permissions).

* **Dynatrace Setup**

  The target VKS cluster must have the Dynatrace OneAgent installed.

* In Harness, the Service appdeployservice must have a Health Source pointing to your Dynatrace dashboard/API.

### Validation & Monitoring

**STO Reporting**

Navigate to the Security Tests tab in Harness to view the vulnerabilities found in the Repo and Container scans.

**SRM Reliability Tracking**

  * Metric Trends: Visualizes how the deployment impacted application health. Harness pulls metrics from Dynatrace (CPU, Memory, Error Rates).

  * Automated Rollback: If the SRM analysis fails due to Dynatrace reporting high error rates, pipeline trggers a safe rollback automatically.


## Reference

- [Harness STO](https://developer.harness.io/docs/security-testing-orchestration/overview/)
- [Harness SRM](https://developer.harness.io/docs/service-reliability-management/)
- [Wiz Scanner](https://www.wiz.io/integrations/harness)
- [Dynatrace Health]()

