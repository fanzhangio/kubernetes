# Deploy Kubelet with allowed-numa-nodes Feature on GB200

## Prerequisites
- GB200 Grace-Hopper system (ARM64/aarch64)
- Root/sudo access
- Go 1.21+ installed
- Existing Kubernetes cluster with kubelet running

## Step 1: Build Kubelet on GB200

On your GB200 system, run:

```bash
# Clone and build
git clone https://github.com/fanzhang/kubernetes.git
cd kubernetes
git checkout topology-numa-dev

# Build ARM64 kubelet
make WHAT=cmd/kubelet KUBE_BUILD_PLATFORMS=linux/arm64

# Verify build
ls -lh _output/local/bin/linux/arm64/kubelet
```

## Step 2: Backup Current Kubelet

```bash
# Backup existing kubelet
sudo cp /usr/bin/kubelet /usr/bin/kubelet.backup.$(date +%Y%m%d)

# Check current version
kubelet --version
```

## Step 3: Stop Kubelet Service

```bash
# Stop kubelet
sudo systemctl stop kubelet

# Verify it's stopped
sudo systemctl status kubelet
```

## Step 4: Update Kubelet Configuration

Edit your kubelet config file (usually `/var/lib/kubelet/config.yaml`):

```yaml
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration

# ... keep all your existing settings ...

topologyManagerPolicy: best-effort
topologyManagerPolicyOptions:
  prefer-closest-numa-nodes: "true"
  allowed-numa-nodes: "0,1"  # NEW: Only use NUMA nodes 0 and 1

cpuManagerPolicy: static
cpuManagerPolicyOptions:
  distribute-cpus-across-numa: "true"

# Remove this line if present:
# max-allowable-numa-nodes: "30"

kubeReserved:
  cpu: 4000m
  memory: 10Gi
  ephemeral-storage: 20Gi
systemReserved:
  cpu: 4000m
  memory: 10Gi
  ephemeral-storage: 20Gi
```

**Important:** Update the config file:

```bash
sudo vim /var/lib/kubelet/config.yaml
# or
sudo nano /var/lib/kubelet/config.yaml
```

## Step 5: Replace Kubelet Binary

```bash
# Copy new kubelet
sudo cp _output/local/bin/linux/arm64/kubelet /usr/bin/kubelet

# Set proper permissions
sudo chmod +x /usr/bin/kubelet

# Verify new version
/usr/bin/kubelet --version
```

## Step 6: Clean State Files (Important!)

The CPU and Topology managers maintain state files that need to be cleared when changing configuration:

```bash
# Stop kubelet if not already stopped
sudo systemctl stop kubelet

# Remove state files
sudo rm -f /var/lib/kubelet/cpu_manager_state
sudo rm -f /var/lib/kubelet/memory_manager_state

# Optional: Remove device plugin state
sudo rm -f /var/lib/kubelet/device-plugins/kubelet_internal_checkpoint
```

## Step 7: Start Kubelet

```bash
# Start kubelet
sudo systemctl start kubelet

# Check status
sudo systemctl status kubelet

# Follow logs
sudo journalctl -u kubelet -f
```

## Step 8: Verify the Feature is Working

Check kubelet logs for the new feature:

```bash
# Look for the allowed-numa-nodes log message
sudo journalctl -u kubelet | grep -i "allowed-numa-nodes"

# Should see something like:
# "Filtering NUMA nodes to allowed list" allowed-numa-nodes=[0,1]
```

Check that topology manager initialized correctly:

```bash
# Look for topology manager initialization
sudo journalctl -u kubelet | grep -i "topology manager"

# Should NOT see errors about "unsupported on machines with more than X NUMA Nodes"
```

Verify NUMA topology:

```bash
# Check that kubelet sees only 2 NUMA nodes now
kubectl get --raw /api/v1/nodes/$(hostname)/proxy/configz | jq '.kubeletconfig.topologyManagerPolicyOptions'
```

## Step 9: Test with a Pod

Create a test pod that requires topology alignment:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-topology-numa
spec:
  containers:
  - name: test
    image: ubuntu:22.04
    command: ["sleep", "infinity"]
    resources:
      requests:
        cpu: "4"
        memory: "8Gi"
      limits:
        cpu: "4"
        memory: "8Gi"
```

Apply and check:

```bash
kubectl apply -f test-pod.yaml

# Check pod status
kubectl get pod test-topology-numa

# Check which NUMA nodes were assigned
kubectl exec test-topology-numa -- numactl -H
kubectl exec test-topology-numa -- cat /proc/self/status | grep Cpus_allowed_list
```

## Troubleshooting

### Issue: Kubelet won't start

```bash
# Check logs
sudo journalctl -u kubelet -n 100 --no-pager

# Common issues:
# 1. Config file syntax error - validate YAML
# 2. Feature gate not enabled - check featureGates section
# 3. State files conflict - remove state files as shown in Step 6
```

### Issue: "Invalid NUMA node" errors

```bash
# Verify your NUMA topology
lscpu | grep NUMA
numactl -H

# Make sure allowed-numa-nodes only lists nodes that have CPUs
# For GB200: nodes 0,1 have CPUs, nodes 2-33 are empty/GPU memory
```

### Issue: Pods not getting scheduled

```bash
# Check node allocatable resources
kubectl describe node $(hostname)

# Check topology manager admission
sudo journalctl -u kubelet | grep -i "topology.*admit"
```

### Rollback if Needed

```bash
# Stop kubelet
sudo systemctl stop kubelet

# Restore backup
sudo cp /usr/bin/kubelet.backup.$(date +%Y%m%d) /usr/bin/kubelet

# Restore old config
sudo cp /var/lib/kubelet/config.yaml.backup /var/lib/kubelet/config.yaml

# Clean state
sudo rm -f /var/lib/kubelet/cpu_manager_state
sudo rm -f /var/lib/kubelet/memory_manager_state

# Start kubelet
sudo systemctl start kubelet
```

## Verification Checklist

- ✅ Kubelet starts successfully
- ✅ Logs show "Filtering NUMA nodes to allowed list" with [0,1]
- ✅ No errors about "unsupported on machines with more than X NUMA Nodes"
- ✅ Node shows as Ready in `kubectl get nodes`
- ✅ Test pod can be scheduled and runs successfully
- ✅ CPU manager shows CPUs only from NUMA nodes 0 and 1

## Performance Monitoring

After deployment, monitor:

```bash
# CPU allocation
kubectl describe node $(hostname) | grep -A 20 "Allocated resources"

# NUMA alignment of pods
for pod in $(kubectl get pods -o name); do
  echo "=== $pod ==="
  kubectl exec $pod -- numactl -H 2>/dev/null || true
done

# Topology manager metrics (if prometheus is configured)
kubectl get --raw /metrics | grep topology_manager
```
