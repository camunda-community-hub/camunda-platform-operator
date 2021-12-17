/*
Copyright 2021.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ZeebeSpec defines the desired state of Zeebe
type ZeebeSpec struct {
	// Broker configurations
	Broker BrokerSpec `json:"broker,omitempty"`

	// Gateway configurations
	Gateway GatewaySpec `json:"gateway,omitempty"`
}

type BrokerSpec struct {
	Partitions PartitionsSpec `json:"partitions,omitempty"`
	Backend    BackendSpec    `json:"backend,omitempty"`
}

type PartitionsSpec struct {
	// how many partitions the cluster should have
	Count *int8 `json:"count,omitempty"`

	// how often a partition should be replicated
	Replication *int `json:"replication,omitempty"`
}

type GatewaySpec struct {
	// per default false, which means we use an embedded gateway
	// +optional
	Standalone *bool `json:"standalone,omitempty"`
	// Optional, only necessary if the gateway is standalone
	// +optional
	Backend BackendSpec `json:"backend,omitempty"`
}

type BackendSpec struct {
	// Repository and name of the container image to use
	// +optional
	ImageName string `json:"imageName,omitempty"`
	// Tag the container image to use. Tags matching /snapshot/i will use ImagePullPolicy Always
	// +optional
	ImageTag string `json:"imageTag,omitempty"`

	// Resources which should be used by the component
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Any var set here will override those provided to the container.
	// Behaviour if duplicate vars are provided _here_ is undefined.
	OverrideEnv []v1.EnvVar `json:"overrideEnv,omitempty"`

	// The replication count for the component
	Replicas *int8 `json:"replicas,omitempty"`
}

// ZeebeStatus defines the observed state of Zeebe
type ZeebeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PodName of the active node.
	Active string `json:"active"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Zeebe is the Schema for the zeebes API
type Zeebe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZeebeSpec   `json:"spec,omitempty"`
	Status ZeebeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZeebeList contains a list of Zeebe
type ZeebeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zeebe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zeebe{}, &ZeebeList{})
}
