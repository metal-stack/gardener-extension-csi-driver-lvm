package csidriverlvm

import (
	corev1 "k8s.io/api/core/v1"
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

	// DefaultStorageClass can be set to a name of a storage class deployed by this extension, which will then be marked as the default storage class.
	DefaultStorageClass *string

	// PullPolicy can be set to adjust the pull policy of the deployed components (development purpose)
	PullPolicy *corev1.PullPolicy

	// CapacityPollInterval can be set to determine the poll interval of the capacity tracking  (default: 5m)
	CapacityPollInterval *string
}
