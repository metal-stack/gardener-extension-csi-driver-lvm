package csidriverlvm

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CsiDriverLvmConfig configuration resource
type CsiDriverLvmConfig struct {
	metav1.TypeMeta
}
