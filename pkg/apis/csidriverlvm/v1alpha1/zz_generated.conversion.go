//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
2025 Copyright metal-stack Authors.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	csidriverlvm "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/csidriverlvm"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*CsiDriverLvmConfig)(nil), (*csidriverlvm.CsiDriverLvmConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_CsiDriverLvmConfig_To_csidriverlvm_CsiDriverLvmConfig(a.(*CsiDriverLvmConfig), b.(*csidriverlvm.CsiDriverLvmConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*csidriverlvm.CsiDriverLvmConfig)(nil), (*CsiDriverLvmConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_csidriverlvm_CsiDriverLvmConfig_To_v1alpha1_CsiDriverLvmConfig(a.(*csidriverlvm.CsiDriverLvmConfig), b.(*CsiDriverLvmConfig), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_CsiDriverLvmConfig_To_csidriverlvm_CsiDriverLvmConfig(in *CsiDriverLvmConfig, out *csidriverlvm.CsiDriverLvmConfig, s conversion.Scope) error {
	out.DevicePattern = (*string)(unsafe.Pointer(in.DevicePattern))
	out.HostWritePath = (*string)(unsafe.Pointer(in.HostWritePath))
	out.DefaultStorageClass = (*string)(unsafe.Pointer(in.DefaultStorageClass))
	return nil
}

// Convert_v1alpha1_CsiDriverLvmConfig_To_csidriverlvm_CsiDriverLvmConfig is an autogenerated conversion function.
func Convert_v1alpha1_CsiDriverLvmConfig_To_csidriverlvm_CsiDriverLvmConfig(in *CsiDriverLvmConfig, out *csidriverlvm.CsiDriverLvmConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_CsiDriverLvmConfig_To_csidriverlvm_CsiDriverLvmConfig(in, out, s)
}

func autoConvert_csidriverlvm_CsiDriverLvmConfig_To_v1alpha1_CsiDriverLvmConfig(in *csidriverlvm.CsiDriverLvmConfig, out *CsiDriverLvmConfig, s conversion.Scope) error {
	out.DevicePattern = (*string)(unsafe.Pointer(in.DevicePattern))
	out.HostWritePath = (*string)(unsafe.Pointer(in.HostWritePath))
	out.DefaultStorageClass = (*string)(unsafe.Pointer(in.DefaultStorageClass))
	return nil
}

// Convert_csidriverlvm_CsiDriverLvmConfig_To_v1alpha1_CsiDriverLvmConfig is an autogenerated conversion function.
func Convert_csidriverlvm_CsiDriverLvmConfig_To_v1alpha1_CsiDriverLvmConfig(in *csidriverlvm.CsiDriverLvmConfig, out *CsiDriverLvmConfig, s conversion.Scope) error {
	return autoConvert_csidriverlvm_CsiDriverLvmConfig_To_v1alpha1_CsiDriverLvmConfig(in, out, s)
}
