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

func (config *CsiDriverLvmConfig) ConfigureDefaults(hostWritePath *string, devicePattern *string) {
	if config.HostWritePath == nil {
		config.HostWritePath = hostWritePath
	}
	if config.DevicePattern == nil {
		config.DevicePattern = devicePattern
	}
}

func (config *CsiDriverLvmConfig) IsValid() bool {
	re := regexp.MustCompile(`^(/[^/ ]*)+/?$`)

	if (config.HostWritePath == nil) || (config.DevicePattern == nil) {
		println("HostWritePath or DevicePattern is nil", config.HostWritePath, config.DevicePattern)
		return false
	}

	return re.MatchString(*config.HostWritePath) && re.MatchString(*config.DevicePattern)
}
