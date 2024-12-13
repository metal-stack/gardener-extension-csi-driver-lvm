package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	healthcheckconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the csi-driver-lvm controller.
type ControllerConfiguration struct {
	metav1.TypeMeta

	// DefaultDevicePattern can be used to configure the glob pattern for the devices used by the LVM driver
	DefaultDevicePattern *string

	// DefaultHostWritePath can be used to configure the default path for the host write path - used on read-only filesystems (Talos  OS "/var/etc/lvm")
	DefaultHostWritePath *string

	// HealthCheckConfig is the config for the health check controller
	HealthCheckConfig *healthcheckconfig.HealthCheckConfig
}
