package v1alpha1

import (
	"path/filepath"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	ShootCsiDriverLvmResourceName = "extension-csi-driver-lvm"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration configuration resource
type CsiDriverLvmConfig struct {
	metav1.TypeMeta `json:",inline"`

	// DevicePattern can be used to configure the glob pattern for the devices used by the LVM driver
	// +optional
	DevicePattern *string `json:"devicePattern,omitempty"`

	// HostWritePath can be used to configure the host write path - used on read-only filesystems (Talos  OS "/var/etc/lvm")
	// +optional
	HostWritePath *string `json:"hostWritePath,omitempty"`

	// DefaultStorageClass can be set to a name of a storage class deployed by this extension, which will then be marked as the default storage class.
	// +optional
	DefaultStorageClass *string `json:"defaultStorageClass,omitempty"`

	// PullPolicy can be set to adjust the pull policy of the deployed components (development purpose). Defaults to "IfNotPresent".
	// +optional
	PullPolicy *corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// Encryption enables encrypted StorageClass variants (linear-encrypted,
	// mirror-encrypted, striped-encrypted). When set, the extension creates
	// additional StorageClasses that reference a user-provided LUKS key Secret
	// living in the shoot cluster. The user is responsible for creating and
	// rotating that Secret; the extension never reads or manages its contents.
	// +optional
	Encryption *EncryptionConfig `json:"encryption,omitempty"`
}

// EncryptionConfig configures LUKS-encrypted StorageClass variants.
type EncryptionConfig struct {
	// SecretRef points to a Secret in the shoot cluster holding the LUKS key
	// material. Both Name and Namespace are required.
	SecretRef corev1.SecretReference `json:"secretRef"`
}

func (config *CsiDriverLvmConfig) ConfigureDefaults(hostWritePath *string, devicePattern *string) {
	if config.HostWritePath == nil {
		config.HostWritePath = hostWritePath
	}
	if config.DevicePattern == nil {
		config.DevicePattern = devicePattern
	}
	if config.PullPolicy == nil {
		config.PullPolicy = ptr.To(corev1.PullIfNotPresent)
	}
}

func (config *CsiDriverLvmConfig) IsValid(log logr.Logger) bool {
	if (config.HostWritePath == nil) || (config.DevicePattern == nil) {
		log.Info("hostWritePath or devicePattern is nil", config.HostWritePath, config.DevicePattern)
		return false
	}

	if *config.HostWritePath == "" || *config.DevicePattern == "" {
		log.Info("hostWritePath or devicePattern is empty", config.HostWritePath, config.DevicePattern)
		return false
	}

	//glob pattern validation could be problematic -> go glob interpretation can be different from bash
	_, err := filepath.Match(*config.DevicePattern, "")
	if err != nil {
		log.Info("bad device pattern")
		return false
	}

	hasValidHostWritePath := filepath.IsAbs(*config.HostWritePath)
	if !hasValidHostWritePath {
		log.Info("hostWritePath is not absolute")
		return false
	}

	if config.Encryption != nil {
		if config.Encryption.SecretRef.Name == "" || config.Encryption.SecretRef.Namespace == "" {
			log.Info("encryption.secretRef.name and encryption.secretRef.namespace must both be set")
			return false
		}
	}

	return true
}
