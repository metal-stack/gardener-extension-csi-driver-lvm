package v1alpha1

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

var log = logr.New(logr.Discard().GetSink())

func stringPtr(s string) *string {
	return &s
}

func TestConfig(t *testing.T) {

	tt := []struct {
		desc       string
		customData *CsiDriverLvmConfig
		valid      bool
	}{
		{
			desc: "test nil config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: nil,
				HostWritePath: nil,
			},
			valid: false,
		},
		{
			desc: "test devicePattern nil config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: nil,
				HostWritePath: stringPtr("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test hostWritePath nil config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("/dev/loop100"),
				HostWritePath: nil,
			},
			valid: false,
		},
		{
			desc: "test empty config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr(""),
				HostWritePath: stringPtr(""),
			},
			valid: false,
		},
		{
			desc: "test empty devicePattern config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr(""),
				HostWritePath: stringPtr("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test empty hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("/dev/loop1"),
				HostWritePath: stringPtr(""),
			},
			valid: false,
		},
		{
			desc: "test invalid devicePattern config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("[a-"),
				HostWritePath: stringPtr("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test not absolute hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("[a-z]"),
				HostWritePath: stringPtr("./etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test not absolute hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("[a-z]"),
				HostWritePath: stringPtr("etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test valid config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: stringPtr("/dev/loop10[0,1]"),
				HostWritePath: stringPtr("/etc/lvm"),
			},
			valid: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			println(tc.desc)
			isConfigValid := tc.customData.IsValid(log)
			assert.Equal(t, tc.valid, isConfigValid)
		})
	}
}
