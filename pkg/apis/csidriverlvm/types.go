package csidriverlvm

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CsiDriverLvmConfig configuration resource
type CsiDriverLvmConfig struct {
	metav1.TypeMeta

	// DevicePattern can be used to configure the glob pattern for the devices used by the LVM driver
	DevicePattern *string

	// HostWritePath can be used to configure the host write path - used on read-only filesystems (Talos  OS "/var/etc/lvm")
	HostWritePath *string
}
