package v1alpha1

import (
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ShootCsiDriverLvmResourceName = "extension-csi-driver-lvm"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration configuration resource
type CsiDriverLvmConfig struct {
	metav1.TypeMeta `json:",inline"`

	DevicePattern *string `json:"devicePattern,omitempty"`
	HostWritePath *string `json:"hostWritePath,omitempty"`
}

func (config *CsiDriverLvmConfig) IsVaild() bool {
	re := regexp.MustCompile(`^(/[^/ ]*)+/?$`)

	return re.MatchString(*config.HostWritePath) && re.MatchString(*config.DevicePattern)
}
