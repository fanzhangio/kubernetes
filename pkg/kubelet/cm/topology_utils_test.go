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
	"reflect"
	"testing"

	cadvisorapi "github.com/google/cadvisor/info/v1"
)

func TestFilterNUMATopology(t *testing.T) {
	topology := []cadvisorapi.Node{
		{Id: 0, Memory: 64 * 1024 * 1024 * 1024},
		{Id: 1, Memory: 64 * 1024 * 1024 * 1024},
		{Id: 2, Memory: 0},
		{Id: 3, Memory: 0},
		{Id: 4, Memory: 0},
	}

	tests := []struct {
		name        string
		excluded    []int
		expectedIDs []int
	}{
		{
			name:        "no exclusions returns original topology",
			excluded:    nil,
			expectedIDs: []int{0, 1, 2, 3, 4},
		},
		{
			name:        "exclude zero-memory nodes",
			excluded:    []int{2, 3, 4},
			expectedIDs: []int{0, 1},
		},
		{
			name:        "exclude subset of nodes",
			excluded:    []int{3},
			expectedIDs: []int{0, 1, 2, 4},
		},
		{
			name:        "exclude all nodes",
			excluded:    []int{0, 1, 2, 3, 4},
			expectedIDs: []int{},
		},
		{
			name:        "exclude non-existent node IDs is harmless",
			excluded:    []int{99, 100},
			expectedIDs: []int{0, 1, 2, 3, 4},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filtered := filterNUMATopology(topology, tc.excluded)
			gotIDs := make([]int, len(filtered))
			for i, n := range filtered {
				gotIDs[i] = n.Id
			}
			if len(gotIDs) != len(tc.expectedIDs) {
				t.Fatalf("expected %v, got %v", tc.expectedIDs, gotIDs)
			}
			for i := range gotIDs {
				if gotIDs[i] != tc.expectedIDs[i] {
					t.Fatalf("expected %v, got %v", tc.expectedIDs, gotIDs)
				}
			}
		})
	}
}

func TestGetZeroMemoryNUMANodes(t *testing.T) {
	tests := []struct {
		name     string
		topology []cadvisorapi.Node
		expected []int
	}{
		{
			name: "GB200-like: CPU nodes with memory, many zero-memory nodes",
			topology: []cadvisorapi.Node{
				{Id: 0, Memory: 64 * 1024 * 1024 * 1024},
				{Id: 1, Memory: 64 * 1024 * 1024 * 1024},
				{Id: 2, Memory: 128 * 1024 * 1024 * 1024},
				{Id: 3, Memory: 0},
				{Id: 4, Memory: 0},
				{Id: 5, Memory: 0},
			},
			expected: []int{3, 4, 5},
		},
		{
			name: "all nodes have memory",
			topology: []cadvisorapi.Node{
				{Id: 0, Memory: 32 * 1024 * 1024 * 1024},
				{Id: 1, Memory: 32 * 1024 * 1024 * 1024},
			},
			expected: nil,
		},
		{
			name: "node with zero regular memory but hugepages is not zero-memory",
			topology: []cadvisorapi.Node{
				{Id: 0, Memory: 64 * 1024 * 1024 * 1024},
				{Id: 1, Memory: 0, HugePages: []cadvisorapi.HugePagesInfo{
					{PageSize: 1048576, NumPages: 10},
				}},
				{Id: 2, Memory: 0},
			},
			expected: []int{2},
		},
		{
			name:     "empty topology",
			topology: nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getZeroMemoryNUMANodes(tc.topology)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
