#!/bin/bash
# Build and Deploy Kubelet with allowed-numa-nodes Feature
# Run this script on your GB200 (Grace-Hopper) ARM64 system

set -e

echo "=== Building ARM64 Kubelet with allowed-numa-nodes feature ==="

# Clone your fork if not already done
if [ ! -d "kubernetes" ]; then
    echo "Cloning repository..."
    git clone https://github.com/fanzhang/kubernetes.git
    cd kubernetes
    git checkout topology-numa-dev
else
    echo "Repository exists, updating..."
    cd kubernetes
    git fetch origin
    git checkout topology-numa-dev
    git pull origin topology-numa-dev
fi

# Build kubelet for ARM64
echo "Building kubelet..."
make WHAT=cmd/kubelet KUBE_BUILD_PLATFORMS=linux/arm64

# Check the build
if [ -f "_output/local/bin/linux/arm64/kubelet" ]; then
    echo "✅ Kubelet built successfully!"
    ls -lh _output/local/bin/linux/arm64/kubelet
else
    echo "❌ Build failed!"
    exit 1
fi

echo ""
echo "=== Build Complete ==="
echo "Kubelet binary location: $(pwd)/_output/local/bin/linux/arm64/kubelet"
echo ""
echo "Next steps:"
echo "1. Backup current kubelet: sudo cp /usr/bin/kubelet /usr/bin/kubelet.backup"
echo "2. Stop kubelet: sudo systemctl stop kubelet"
echo "3. Replace kubelet: sudo cp _output/local/bin/linux/arm64/kubelet /usr/bin/kubelet"
echo "4. Update config: /var/lib/kubelet/config.yaml"
echo "5. Start kubelet: sudo systemctl start kubelet"
