// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addPytorchDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// setPytorchDefaultPort sets the default ports for pytorch container.
func setPytorchDefaultPort(spec *corev1.PodSpec) {
	index := getDefaultContainerIndex(spec, PytorchJobDefaultContainerName)
	if ok := hasDefaultPort(spec, index, PytorchJobDefaultPortName); !ok {
		setDefaultPort(spec, PytorchJobDefaultPortName, PytorchJobDefaultPort, index)
	}
}

func setElasticPolicy(pytorchJob *PyTorchJob) {
	if pytorchJob.Spec.ElasticPolicy != nil {
		if pytorchJob.Spec.ElasticPolicy.MaxReplicas != nil &&
			pytorchJob.Spec.ElasticPolicy.MinReplicas != nil {
			return
		} else if pytorchJob.Spec.ElasticPolicy.MaxReplicas != nil {
			// Set MinRepliacs to elasticPolicy.MaxReplicas.
			pytorchJob.Spec.ElasticPolicy.MinReplicas = pytorchJob.Spec.ElasticPolicy.MaxReplicas
		} else if pytorchJob.Spec.ElasticPolicy.MinReplicas != nil {
			pytorchJob.Spec.ElasticPolicy.MaxReplicas = pytorchJob.Spec.ElasticPolicy.MinReplicas
		} else {
			workerReplicas := pytorchJob.Spec.PyTorchReplicaSpecs[PyTorchJobReplicaTypeWorker].Replicas
			// Set Min and Max to worker.spec.Replicas.
			pytorchJob.Spec.ElasticPolicy.MaxReplicas = workerReplicas
			pytorchJob.Spec.ElasticPolicy.MinReplicas = workerReplicas
		}
	}
}

// setPytorchTypeNamesToCamelCase sets the name of all replica types from any case to correct case.
func setPytorchTypeNamesToCamelCase(pytorchJob *PyTorchJob) {
	replicaTypes := []ReplicaType{
		PyTorchJobReplicaTypeMaster,
		PyTorchJobReplicaTypeWorker,
	}
	for _, replicaType := range replicaTypes {
		setTypeNameToCamelCase(pytorchJob.Spec.PyTorchReplicaSpecs, replicaType)
	}
}

// SetDefaults_PyTorchJob sets any unspecified values to defaults.
func SetDefaults_PyTorchJob(job *PyTorchJob) {
	// Set default cleanpod policy to None.
	if job.Spec.RunPolicy.CleanPodPolicy == nil {
		policy := CleanPodPolicyNone
		job.Spec.RunPolicy.CleanPodPolicy = &policy
	}

	// Update the key of PyTorchReplicaSpecs to camel case.
	setPytorchTypeNamesToCamelCase(job)

	for _, spec := range job.Spec.PyTorchReplicaSpecs {
		setDefaultReplicas(spec, 1)
		setDefaultRestartPolicy(spec, PytorchJobDefaultRestartPolicy)
		setPytorchDefaultPort(&spec.Template.Spec)
	}
	// Set default elastic policy.
	setElasticPolicy(job)
}
