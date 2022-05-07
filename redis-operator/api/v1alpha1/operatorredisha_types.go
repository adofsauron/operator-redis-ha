/*
Copyright 2022.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubernetesConfig will be the JSON struct for Basic Redis Config
type KubernetesConfig struct {
	Namespapce      string                       `json:"namespace"`
	Image           string                       `json:"image"`
	ImagePullPolicy corev1.PullPolicy            `json:"imagePullPolicy,omitempty"`
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// Storage is the inteface to add pvc and pv support in redis
type Storage struct {
	VolumeClaimTemplate corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// RedisConfig defines the external configuration of Redis
type RedisConfig struct {
	AdditionalRedisConfig *string `json:"additionalRedisConfig,omitempty"`
}

// OperatorRedisHASpec defines the desired state of Redis
type OperatorRedisHASpec struct {
	KubernetesConfig KubernetesConfig     `json:"kubernetesConfig"`
	RedisConfig      *RedisConfig         `json:"redisConfig,omitempty"`
	Storage          *Storage             `json:"storage,omitempty"`
	NodeSelector     map[string]string    `json:"nodeSelector,omitempty"`
	Affinity         *corev1.Affinity     `json:"affinity,omitempty"`
	Tolerations      *[]corev1.Toleration `json:"tolerations,omitempty"`
}

// OperatorRedisHAStatus defines the observed state of OperatorRedisHA
type OperatorRedisHAStatus struct {
	// insert ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CRStatus     int    `json:"crStatus,omitempty"`
	BeSlaveOf    bool   `json:"beSlaveOf,omitempty"`
	RedisAddr    string `json:"redisAddr,omitempty"`
	BeSetEtcdCrt bool   `json:"beSetEtcdCrt,omitempty"`
	RedisPort    int    `json:"redisPort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OperatorRedisHA is the Schema for the operatorredishas API
type OperatorRedisHA struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorRedisHASpec   `json:"spec,omitempty"`
	Status OperatorRedisHAStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OperatorRedisHAList contains a list of OperatorRedisHA
type OperatorRedisHAList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorRedisHA `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorRedisHA{}, &OperatorRedisHAList{})
}
