# Overview

This guide walks you through provisioning a VKS Kubernetes cluster utilizing a custom ClusterClass integrated with Keycloak for OpenID Connect (OIDC) authentication. This integration enables a Harness Kubernetes Connector to authenticate to the VKS cluster using Machine-to-Machine (M2M) access tokens, leveraging Harness Delegates running in Cluster A. Once Authenticated, Harness can deploy applications to OIDC integrated VKS cluster using CI/CD pipelines.

## Prerequisites

* Active administrative access to the vSphere Supervisor Cluster.

* The VCF CLI and associated plugins installed and configured locally.

* Administrative access to your Keycloak instance (to retrieve realm parameters, client configurations, and CA certificates).

## Step-by-Step Deployment Guide

### Step 1: Create the Custom ClusterClass
Switch your command-line context to the vSphere Supervisor cluster and apply the custom ClusterClass in vmware-system-vks-public namespace. This cluster blueprint defines the baseline inline templates and executes the required JSON structural mutations to patch the API server flags.

```bash
# Authenticate and switch context to the Supervisor cluster
kubectl config use-context <SUPERVISOR-CONTEXT>

# Apply the custom ClusterClass definition
kubectl apply -f clusterclass-oidc.yaml
```

### Step 2: Deploy the Authentication Configuration Secret

Before instructing the topology engine to create the workload cluster, you must provision the Secret that holds the structured AuthenticationConfiguration.

> **Note:**  CRITICAL CONFIGURATION RULE: The Kubernetes OIDC schema mandates using the certificateAuthority field containing the raw, unencoded multi-line PEM text block. Do not use certificateAuthorityData or a base64-encoded string, as this will cause a strict schema decoding exception and crash the kube-apiserver static pod.

Apply the configuration secret inside your target namespace:

```bash
kubectl apply -f cluster-oidc-secret.yaml
```

### Step 3: Provision the Workload Cluster
Apply the core Cluster deployment manifest which references your newly created custom ClusterClass.

```bash
kubectl apply -f cluster-oidc.yaml
```

The Cluster API topology controller will process the manifests, match the validation tokens, and instruct vCenter to clone the node templates. 


### Step 4: Map RBAC Roles for Harness CICD Pipelines

Once the VKS workload cluster reaches the Running phase, the cluster administrator must fetch its administrative kubeconfig and bind the incoming Keycloak user groups to target roles within the workload cluster.

Run the following commands to extract the credentials and apply the clusterrolebinding.yaml policy directly onto the newly created VKS instance:

```bash
# Fetch the workload cluster's localized kubeconfig
kubectl get secret <cluster-name>-kubeconfig -n test-namespace -o jsonpath='{.data.value}' | base64 -d > workload.kubeconfig

# Bind the Keycloak groups to the cluster-admin role inside the workload cluster
kubectl --kubeconfig=workload.kubeconfig apply -f clusterrolebinding.yaml
```

This mapping validates the keycloak-group:default-roles-vks-realm token payload sent by the Harness Delegate (running in Cluster A), unblocking continuous application deployment tasks.

## References
- [Custom ClusterClass Example: External OIDC Authentication](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-service-administration-and-development/9-0/managing-vsphere-kuberenetes-service-clusters-and-workloads/provisioning-tkg-service-clusters/using-the-cluster-v1beta1-api/using-the-versioned-clusterclass/example.html#GUID-36bb3bf7-32fb-45ab-889e-c5eb42715b73-en)
