package v1alpha1

import (
	healthcheckconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration configuration resource
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// DefaultDevicePattern can be used to configure the glob pattern for the devices used by the LVM driver
	// +optional
	DefaultDevicePattern *string `json:"defaultDevicePattern,omitempty"`

	// DefaultHostWritePath can be used to configure the default path for the host write path - used on read-only filesystems (Talos  OS "/var/etc/lvm")
	// +optional
	DefaultHostWritePath *string `json:"defaultHostWritePath,omitempty"`

	// HealthCheckConfig is the config for the health check controller
	// +optional
	HealthCheckConfig *healthcheckconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`
}
