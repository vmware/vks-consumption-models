# Example Client VM Configuration & Tooling #

Based on Ubuntu Jammy. The cloud OVA image is available at: 
https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.ova 

Connecting to Supervisor and VKS Clusters:
https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/configuring-identity-and-access-for-tkg-service-clusters/connecting-to-vsphere-with-tanzu-clusters.html

VCF CLI command reference:
https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-9-0-and-later/9-0/building-your-cloud-applications/getting-started-with-the-tools-for-building-applications/installing-and-using-vcf-cli-v9/command-reference2.html


## Install VCF Command Line
```
# Download VCF CLI & install (version 9.0)
curl -fsSL https://packages.broadcom.com/artifactory/vcf-distro\
  /vcf-cli/linux/amd64/v9.0.0/vcf-cli.tar.gz | tar xz

sudo install -m 755 vcf-cli-linux_amd64 /usr/local/bin/vcf

# (Optional) Add autocomplete for VCF command line (bash shell)
echo "source <(vcf completion bash)" >> ~/.bashrc

# Create supervisor context & login
vcf context create --endpoint=<supervisor endpoint> \
  --username administrator@vsphere.local

# Use supervisor context
vcf context use supervisor-namespace

# Refresh context to re-auth
vcf context refresh supervisor-namespace

# Create VKS workload cluster context
vcf context create --endpoint=<supervisor endpoint> \
  --username administrator@vsphere.local \
  --workload-cluster-name my-vks-cluster \
  --workload-cluster-namespace supervisor-namespace
```

## Install vCenter Certificate
```
# Get vCenter certs & install
# Download the zip file to /tmp using curl (insecure mode required)
VCENTER_IP=<vCenter IP>
curl -k -fsSL -o /tmp/vccert.zip https://${VCENTER_IP}/certs/download.zip

# Unzip and copy to SSL directory
unzip /tmp/vccert.zip -d /tmp
sudo cp /tmp/certs/lin/* /etc/ssl/certs

# Update system certs
sudo update-ca-certificates

```

## Install kubectl
```
sudo apt update
sudo apt install -y kubectl

# Add autocomplete and shorthand 'k' for Kubectl (bash shell)
cat << 'EOF' >> ~/.bashrc
if command -v kubectl &> /dev/null; then 
    source <(kubectl completion bash)
    alias k=kubectl 
    complete -o default -F __start_kubectl k
fi
EOF
source ~/.bashrc
```

## Install Helm. See https://helm.sh/docs/intro/install/
```
# Get Helm using download script
curl -fsSL -o /tmp/get_helm.sh \
  https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-4
chmod 700 get_helm.sh
./get_helm.sh  
```


