# Install VKS Add-on Example (Istio) #

VKS Add-ons reference:
https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/managing-add-ons-in-vks-clusters.html

VKS Istio package reference:
https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/installing-standard-packages-on-tkg-service-clusters/standard-package-reference/istio-package-reference.html

```
# Change context to the Supervisor context, for example:
vcf context use supervisor-namespace

# List the available VKS clusters
vcf cluster list -A

# Create an install task
vcf addon install create istio --cluster-name my_vks_cluster -y

# Change context to the VKS cluster context
vcf context use my_vks_cluster

# Verify installation of Istio
kubectl -n vmware-system-tkg describe $(kubectl get pkgi -A -o name | grep istio)
```
