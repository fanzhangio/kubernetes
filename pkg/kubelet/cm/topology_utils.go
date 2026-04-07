/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cm

import (
	"sort"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func filterNUMATopology(topology []cadvisorapi.Node, excludedNUMANodes []int) []cadvisorapi.Node {
	if len(excludedNUMANodes) == 0 {
		return topology
	}

	excludedSet := sets.New(excludedNUMANodes...)
	filtered := make([]cadvisorapi.Node, 0, len(topology))
	for _, node := range topology {
		if !excludedSet.Has(node.Id) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// getZeroMemoryNUMANodes returns NUMA node IDs that have zero allocatable
// memory (regular + hugepages).  These nodes carry no resources workloads
// can use and should be excluded from topology hint iteration regardless
// of the memory-manager policy to avoid O(2^n) combinatorial blowup on
// large-NUMA platforms such as the NVIDIA GB200 (34 OS-visible NUMA nodes,
// most with zero memory).
func getZeroMemoryNUMANodes(topology []cadvisorapi.Node) []int {
	var zeroNodes []int
	for _, node := range topology {
		if nodeHasZeroMemory(node) {
			zeroNodes = append(zeroNodes, node.Id)
		}
	}
	sort.Ints(zeroNodes)
	return zeroNodes
}

func nodeHasZeroMemory(node cadvisorapi.Node) bool {
	if node.Memory > 0 {
		return false
	}
	for _, hp := range node.HugePages {
		if hp.NumPages > 0 {
			return false
		}
	}
	return true
}
