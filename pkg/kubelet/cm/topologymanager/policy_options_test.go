/*
Copyright 2022 The Kubernetes Authors.

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

package topologymanager

import (
	"fmt"
	"strings"
	"testing"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	pkgfeatures "k8s.io/kubernetes/pkg/features"
)

var fancyBetaOption = "fancy-new-option"
var fancyAlphaOption = "fancy-alpha-option"

type optionAvailTest struct {
	option            string
	featureGate       featuregate.Feature
	featureGateEnable bool
	expectedAvailable bool
}

func TestNewTopologyManagerOptions(t *testing.T) {
	testCases := []struct {
		description       string
		policyOptions     map[string]string
		featureGate       featuregate.Feature
		featureGateEnable bool
		expectedErr       error
		expectedOptions   PolicyOptions
	}{
		{
			description: "return TopologyManagerOptions with PreferClosestNUMA set to true",
			expectedOptions: PolicyOptions{
				PreferClosestNUMA:     true,
				MaxAllowableNUMANodes: 8,
			},
			policyOptions: map[string]string{
				PreferClosestNUMANodes: "true",
				MaxAllowableNUMANodes:  "8",
			},
		},
		{
			description:       "return TopologyManagerOptions with MaxAllowableNUMANodes set to 12",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			expectedOptions: PolicyOptions{
				MaxAllowableNUMANodes: 12,
			},
			policyOptions: map[string]string{
				MaxAllowableNUMANodes: "12",
			},
		},
		{
			description: "fail to set option when TopologyManagerPolicyBetaOptions feature gate is not set",
			featureGate: pkgfeatures.TopologyManagerPolicyBetaOptions,
			policyOptions: map[string]string{
				MaxAllowableNUMANodes: "8",
			},
			expectedErr: fmt.Errorf("topology manager policy beta-level options not enabled,"),
		},
		{
			description: "return empty TopologyManagerOptions",
			expectedOptions: PolicyOptions{
				MaxAllowableNUMANodes: 8,
			},
		},
		{
			description:       "fail to parse options with error PreferClosestNUMANodes",
			featureGateEnable: true,
			policyOptions: map[string]string{
				PreferClosestNUMANodes: "not a boolean",
			},
			expectedErr: fmt.Errorf("bad value for option"),
		},
		{
			description:       "fail to parse options with error MaxAllowableNUMANodes",
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: true,
			policyOptions: map[string]string{
				MaxAllowableNUMANodes: "can't parse to int",
			},
			expectedErr: fmt.Errorf("unable to convert policy option to integer"),
		},
		{
			description:       "test beta options success",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			policyOptions: map[string]string{
				fancyBetaOption: "true",
			},
			expectedOptions: PolicyOptions{
				PreferClosestNUMA:     false,
				MaxAllowableNUMANodes: 8,
			},
		},
		{
			description: "test beta options fail",
			featureGate: pkgfeatures.TopologyManagerPolicyBetaOptions,
			policyOptions: map[string]string{
				fancyBetaOption: "true",
			},
			expectedErr: fmt.Errorf("topology manager policy beta-level options not enabled,"),
		},
		{
			description:       "test alpha options success",
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: true,
			policyOptions: map[string]string{
				fancyAlphaOption: "true",
			},
			expectedOptions: PolicyOptions{
				PreferClosestNUMA:     false,
				MaxAllowableNUMANodes: 8,
			},
		},
		{
			description: "test alpha options fail",
			policyOptions: map[string]string{
				fancyAlphaOption: "true",
			},
			expectedErr: fmt.Errorf("topology manager policy alpha-level options not enabled,"),
		},
		{
			description: "return TopologyManagerOptions with AllowedNUMANodes set to 0,1",
			expectedOptions: PolicyOptions{
				MaxAllowableNUMANodes: 8,
				AllowedNUMANodes:      []int{0, 1},
			},
			policyOptions: map[string]string{
				AllowedNUMANodes: "0,1",
			},
		},
		{
			description: "return TopologyManagerOptions with AllowedNUMANodes set to multiple nodes",
			expectedOptions: PolicyOptions{
				MaxAllowableNUMANodes: 8,
				AllowedNUMANodes:      []int{0, 2, 4, 6},
			},
			policyOptions: map[string]string{
				AllowedNUMANodes: "0,2,4,6",
			},
		},
		{
			description: "fail to parse AllowedNUMANodes with invalid node ID",
			policyOptions: map[string]string{
				AllowedNUMANodes: "0,abc",
			},
			expectedErr: fmt.Errorf("invalid NUMA node ID"),
		},
		{
			description: "fail to parse AllowedNUMANodes with negative node ID",
			policyOptions: map[string]string{
				AllowedNUMANodes: "0,-1",
			},
			expectedErr: fmt.Errorf("NUMA node ID must be non-negative"),
		},
		{
			description: "fail to parse AllowedNUMANodes with duplicate node ID",
			policyOptions: map[string]string{
				AllowedNUMANodes: "0,1,0",
			},
			expectedErr: fmt.Errorf("duplicate NUMA node ID"),
		},
		{
			description: "fail to parse AllowedNUMANodes with empty value",
			policyOptions: map[string]string{
				AllowedNUMANodes: "",
			},
			expectedErr: fmt.Errorf("empty value for option"),
		},
	}

	betaOptions.Insert(fancyBetaOption)
	alphaOptions.Insert(fancyAlphaOption)

	for _, tcase := range testCases {
		t.Run(tcase.description, func(t *testing.T) {
			if tcase.featureGate != "" {
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, tcase.featureGate, tcase.featureGateEnable)
			}
			opts, err := NewPolicyOptions(tcase.policyOptions)
			if tcase.expectedErr != nil {
				if !strings.Contains(err.Error(), tcase.expectedErr.Error()) {
					t.Errorf("Unexpected error message. Have: %s, wants %s", err.Error(), tcase.expectedErr.Error())
				}
				return
			}

			// Compare fields individually to handle slice comparison
			if opts.PreferClosestNUMA != tcase.expectedOptions.PreferClosestNUMA {
				t.Errorf("Expected PreferClosestNUMA to equal %v, not %v", tcase.expectedOptions.PreferClosestNUMA, opts.PreferClosestNUMA)
			}
			if opts.MaxAllowableNUMANodes != tcase.expectedOptions.MaxAllowableNUMANodes {
				t.Errorf("Expected MaxAllowableNUMANodes to equal %v, not %v", tcase.expectedOptions.MaxAllowableNUMANodes, opts.MaxAllowableNUMANodes)
			}
			if !intSlicesEqual(opts.AllowedNUMANodes, tcase.expectedOptions.AllowedNUMANodes) {
				t.Errorf("Expected AllowedNUMANodes to equal %v, not %v", tcase.expectedOptions.AllowedNUMANodes, opts.AllowedNUMANodes)
			}
		})
	}
}

func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPolicyDefaultsAvailable(t *testing.T) {
	testCases := []optionAvailTest{
		{
			option:            "this-option-does-not-exist",
			expectedAvailable: false,
		},
		{
			option:            PreferClosestNUMANodes,
			expectedAvailable: true,
		},
		{
			option:            MaxAllowableNUMANodes,
			expectedAvailable: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.option, func(t *testing.T) {
			err := CheckPolicyOptionAvailable(testCase.option)
			isEnabled := (err == nil)
			if isEnabled != testCase.expectedAvailable {
				t.Errorf("option %q available got=%v expected=%v", testCase.option, isEnabled, testCase.expectedAvailable)
			}
		})
	}
}

func TestPolicyOptionsAvailable(t *testing.T) {
	testCases := []optionAvailTest{
		{
			option:            "this-option-does-not-exist",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: false,
			expectedAvailable: false,
		},
		{
			option:            "this-option-does-not-exist",
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			expectedAvailable: false,
		},
		{
			option:            PreferClosestNUMANodes,
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: false,
			expectedAvailable: true,
		},
		{
			option:            PreferClosestNUMANodes,
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: false,
			expectedAvailable: true,
		},
		{
			option:            fancyAlphaOption,
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: true,
			expectedAvailable: true,
		},
		{
			option:            fancyAlphaOption,
			featureGate:       pkgfeatures.TopologyManagerPolicyAlphaOptions,
			featureGateEnable: false,
			expectedAvailable: false,
		},
		{
			option:            fancyBetaOption,
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: true,
			expectedAvailable: true,
		},
		{
			option:            fancyBetaOption,
			featureGate:       pkgfeatures.TopologyManagerPolicyBetaOptions,
			featureGateEnable: false,
			expectedAvailable: false,
		},
	}
	betaOptions.Insert(fancyBetaOption)
	alphaOptions.Insert(fancyAlphaOption)
	for _, testCase := range testCases {
		t.Run(testCase.option, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, testCase.featureGate, testCase.featureGateEnable)
			defer func() {
				// reset feature flag
				featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, testCase.featureGate, !testCase.featureGateEnable)
			}()
			err := CheckPolicyOptionAvailable(testCase.option)
			isEnabled := (err == nil)
			if isEnabled != testCase.expectedAvailable {
				t.Errorf("option %q available got=%v expected=%v", testCase.option, isEnabled, testCase.expectedAvailable)
			}
		})
	}
}
