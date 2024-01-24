/*
Copyright 2023 The KServe Authors.

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

package inferencegraph

import (
	"github.com/google/go-cmp/cmp"
	v1alpha1 "github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
	"github.com/kserve/kserve/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreateInferenceGraphPodSpec(t *testing.T) {
	type args struct {
		graph  *v1alpha1.InferenceGraph
		config *RouterConfig
	}

	routerConfig := RouterConfig{
		Image:         "kserve/router:v0.10.0",
		CpuRequest:    "100m",
		CpuLimit:      "100m",
		MemoryRequest: "100Mi",
		MemoryLimit:   "500Mi",
	}

	routerConfigWithHeaders := RouterConfig{
		Image:         "kserve/router:v0.10.0",
		CpuRequest:    "100m",
		CpuLimit:      "100m",
		MemoryRequest: "100Mi",
		MemoryLimit:   "500Mi",
		Headers: map[string][]string{
			"propagate": {
				"Authorization",
				"Intuit_tid",
			},
		},
	}

	testIGSpecs := map[string]*v1alpha1.InferenceGraph{
		"basic": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "basic-ig",
				Namespace: "basic-ig-namespace",
			},
			Spec: v1alpha1.InferenceGraphSpec{
				Nodes: map[string]v1alpha1.InferenceRouter{
					v1alpha1.GraphRootNodeName: {
						RouterType: v1alpha1.Sequence,
						Steps: []v1alpha1.InferenceStep{
							{
								InferenceTarget: v1alpha1.InferenceTarget{
									ServiceURL: "http://someservice.exmaple.com",
								},
							},
						},
					},
				},
			},
		},
		"withresource": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "resource-ig",
				Namespace: "resource-ig-namespace",
				Annotations: map[string]string{
					"serving.kserve.io/deploymentMode": string(constants.RawDeployment),
				},
			},

			Spec: v1alpha1.InferenceGraphSpec{
				Nodes: map[string]v1alpha1.InferenceRouter{
					v1alpha1.GraphRootNodeName: {
						RouterType: v1alpha1.Sequence,
						Steps: []v1alpha1.InferenceStep{
							{
								InferenceTarget: v1alpha1.InferenceTarget{
									ServiceURL: "http://someservice.exmaple.com",
								},
							},
						},
					},
				},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("100m"),
						v1.ResourceMemory: resource.MustParse("500Mi"),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("100m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			},
		},

		"withenv": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "env-ig",
				Namespace: "env-ig-namespace",
				Annotations: map[string]string{
					"serving.kserve.io/deploymentMode": string(constants.RawDeployment),
				},
			},

			Spec: v1alpha1.InferenceGraphSpec{
				Nodes: map[string]v1alpha1.InferenceRouter{
					v1alpha1.GraphRootNodeName: {
						RouterType: v1alpha1.Sequence,
						Steps: []v1alpha1.InferenceStep{
							{
								InferenceTarget: v1alpha1.InferenceTarget{
									ServiceURL: "http://someservice.exmaple.com",
								},
							},
						},
					},
				},
			},
		},
	}

	expectedPodSpecs := map[string]*v1.PodSpec{
		"basicgraph": {
			Containers: []v1.Container{
				{
					Image: "kserve/router:v0.10.0",
					Name:  "basic-ig",
					Args: []string{
						"--graph-json",
						"{\"nodes\":{\"root\":{\"routerType\":\"Sequence\",\"steps\":[{\"serviceUrl\":\"http://someservice.exmaple.com\"}]}},\"resources\":{}}",
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("500Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
		"basicgraphwithheaders": {
			Containers: []v1.Container{
				{
					Image: "kserve/router:v0.10.0",
					Name:  "basic-ig",
					Args: []string{
						"--graph-json",
						"{\"nodes\":{\"root\":{\"routerType\":\"Sequence\",\"steps\":[{\"serviceUrl\":\"http://someservice.exmaple.com\"}]}},\"resources\":{}}",
					},
					Env: []v1.EnvVar{
						{
							Name:  "PROPAGATE_HEADERS",
							Value: "Authorization,Intuit_tid",
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("500Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
		"withresource": {
			Containers: []v1.Container{
				{
					Image: "kserve/router:v0.10.0",
					Name:  "resource-ig",
					Args: []string{
						"--graph-json",
						"{\"nodes\":{\"root\":{\"routerType\":\"Sequence\",\"steps\":[{\"serviceUrl\":\"http://someservice.exmaple.com\"}]}},\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"500Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}}",
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("500Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
	}

	scenarios := []struct {
		name     string
		args     args
		expected *v1.PodSpec
	}{
		{
			name: "Basic Inference graph",
			args: args{
				graph:  testIGSpecs["basic"],
				config: &routerConfig,
			},
			expected: expectedPodSpecs["basicgraph"],
		},
		{
			name:     "Inference graph with resource requirements",
			args:     args{testIGSpecs["withresource"], &routerConfig},
			expected: expectedPodSpecs["withresource"],
		},
		{
			name: "Inference graph with propagate headers",
			args: args{
				graph:  testIGSpecs["basic"],
				config: &routerConfigWithHeaders,
			},
			expected: expectedPodSpecs["basicgraphwithheaders"],
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			result := createInferenceGraphPodSpec(tt.args.graph, tt.args.config)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Test %q unexpected result (-want +got): %v", t.Name(), diff)
			}
		})
	}
}

func TestConstructGraphObjectMeta(t *testing.T) {
	type args struct {
		name        string
		namespace   string
		annotations map[string]string
		labels      map[string]string
	}

	scenarios := []struct {
		name     string
		args     args
		expected metav1.ObjectMeta
	}{
		{
			name: "Basic Inference graph",
			args: args{
				name:      "basic-ig",
				namespace: "basic-ig-namespace",
			},
			expected: metav1.ObjectMeta{
				Name:      "basic-ig",
				Namespace: "basic-ig-namespace",
				Labels: map[string]string{
					"serving.kserve.io/inferencegraph": "basic-ig",
				},
				Annotations: map[string]string{},
			},
		},
		{
			name: "Inference graph with annotations",
			args: args{
				name:      "basic-ig",
				namespace: "basic-ig-namespace",
				annotations: map[string]string{
					"test": "test",
				},
			},
			expected: metav1.ObjectMeta{
				Name:      "basic-ig",
				Namespace: "basic-ig-namespace",
				Labels: map[string]string{
					"serving.kserve.io/inferencegraph": "basic-ig",
				},
				Annotations: map[string]string{
					"test": "test",
				},
			},
		},
		{
			name: "Inference graph with labels",
			args: args{
				name:      "basic-ig",
				namespace: "basic-ig-namespace",
				labels: map[string]string{
					"test": "test",
				},
			},
			expected: metav1.ObjectMeta{
				Name:      "basic-ig",
				Namespace: "basic-ig-namespace",
				Labels: map[string]string{
					"serving.kserve.io/inferencegraph": "basic-ig",
					"test":                             "test",
				},
				Annotations: map[string]string{},
			},
		},
		{
			name: "Inference graph with annotations and labels",
			args: args{
				name:      "basic-ig",
				namespace: "basic-ig-namespace",
				annotations: map[string]string{
					"test": "test",
				},
				labels: map[string]string{
					"test": "test",
				},
			},
			expected: metav1.ObjectMeta{
				Name:      "basic-ig",
				Namespace: "basic-ig-namespace",
				Labels: map[string]string{
					"serving.kserve.io/inferencegraph": "basic-ig",
					"test":                             "test",
				},
				Annotations: map[string]string{
					"test": "test",
				},
			},
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			result := constructGraphObjectMeta(tt.args.name, tt.args.namespace, tt.args.annotations, tt.args.labels)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Test %q unexpected result (-want +got): %v", t.Name(), diff)
			}
		})
	}
}
