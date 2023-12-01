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
	"encoding/json"
	v1alpha1api "github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
	"github.com/kserve/kserve/pkg/constants"
	"github.com/kserve/kserve/pkg/controller/v1beta1/inferenceservice/reconcilers/raw"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var logger = logf.Log.WithName("InferenceGraphRawDeployer")

/*
This function helps to create core podspec for a given inference graph spec and router configuration
Also propagates headers onto podspec container environment variables.

This function makes sense to be used in raw k8s deployment mode
*/
func createInferenceGraphPodSpec(graph *v1alpha1api.InferenceGraph, config *RouterConfig) *v1.PodSpec {
	bytes, err := json.Marshal(graph.Spec)
	if err != nil {
		return nil
	}

	//Pod spec with 'router container with resource requirements' and 'affinity' as well
	podSpec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  graph.ObjectMeta.Name,
				Image: config.Image,
				Args: []string{
					"--graph-json",
					string(bytes),
				},
				Resources: constructResourceRequirements(*graph, *config),
			},
		},
		Affinity: graph.Spec.Affinity,
	}

	// Only adding this env variable "PROPAGATE_HEADERS" if router's headers config has the key "propagate"
	value, exists := config.Headers["propagate"]
	if exists {
		podSpec.Containers[0].Env = []v1.EnvVar{
			{
				Name:  constants.RouterHeadersPropagateEnvVar,
				Value: strings.Join(value, ","),
			},
		}
	}

	return podSpec
}

/*
A simple utility to create a basic meta object given name and namespace;  Can be extended to accept labels, annotations as well
*/
func constructGraphObjectMeta(name string, namespace string, annotations map[string]string,
	labels map[string]string) metav1.ObjectMeta {

	if annotations == nil {
		annotations = make(map[string]string)
	}

	if labels == nil {
		labels = make(map[string]string)
	}

	labels[constants.InferenceGraphLabel] = name

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
}

/*
Handles bulk of raw deployment logic for Inference graph controller
1. Constructs PodSpec
2. Constructs Meta and Extensionspec
3. Creates a reconciler
4. Set controller referneces
5. Finally reconcile
*/
func handleInferenceGraphRawDeployment(cl client.Client, scheme *runtime.Scheme, graph *v1alpha1api.InferenceGraph, routerConfig *RouterConfig) (ctrl.Result, error) {
	// create desired service object.
	desiredSvc := createInferenceGraphPodSpec(graph, routerConfig)

	objectMeta := constructGraphObjectMeta(graph.ObjectMeta.Name, graph.ObjectMeta.Namespace, graph.ObjectMeta.Annotations, graph.ObjectMeta.Labels)

	//create the reconciler
	reconciler, err := raw.NewRawKubeReconciler(cl, scheme, objectMeta, nil, desiredSvc)

	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "fails to create NewRawKubeReconciler for inference graph")
	}
	//set Deployment Controller
	if err := controllerutil.SetControllerReference(graph, reconciler.Deployment.Deployment, scheme); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "fails to set deployment owner reference for inference graph")
	}
	//set Service Controller
	if err := controllerutil.SetControllerReference(graph, reconciler.Service.Service, scheme); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "fails to set service owner reference for inference graph")
	}

	//set autoscaler Controller
	if err := reconciler.Scaler.Autoscaler.SetControllerReferences(graph, scheme); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "fails to set autoscaler owner references for inference graph")
	}

	//reconcile
	deployment, err := reconciler.Reconcile()
	logger.Info("reconciled:")

	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "fails to reconcile inference graph raw")
	}

	PropagateRawStatus(&graph.Status, deployment, reconciler.URL)

	return ctrl.Result{}, nil
}

/*
PropagateRawStatus Propagates deployment status onto Inference graph status.  In raw deployment mode, deployment available denotes the ready status for IG
*/
func PropagateRawStatus(graphStatus *v1alpha1api.InferenceGraphStatus, deployment *appsv1.Deployment, url *apis.URL) {

	for _, con := range deployment.Status.Conditions {
		if con.Type == appsv1.DeploymentAvailable {
			graphStatus.URL = url

			conditions := []apis.Condition{
				{
					Type:   apis.ConditionReady,
					Status: v1.ConditionTrue,
				},
			}
			graphStatus.SetConditions(conditions)
			logger.Info("status propagated:")
			break
		}
	}
	graphStatus.ObservedGeneration = deployment.Status.ObservedGeneration
}
