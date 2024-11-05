package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SeedCsiDriverLvmResourceName = "extension-csi-driver-lvm"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration configuration resource
type CsiDriverLvmConfig struct {
	metav1.TypeMeta `json:",inline"`
}
