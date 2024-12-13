package v1alpha1

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var log = logr.New(logr.Discard().GetSink())

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
				HostWritePath: ptr.To("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test hostWritePath nil config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("/dev/loop100"),
				HostWritePath: nil,
			},
			valid: false,
		},
		{
			desc: "test empty config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To(""),
				HostWritePath: ptr.To(""),
			},
			valid: false,
		},
		{
			desc: "test empty devicePattern config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To(""),
				HostWritePath: ptr.To("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test empty hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("/dev/loop1"),
				HostWritePath: ptr.To(""),
			},
			valid: false,
		},
		{
			desc: "test invalid devicePattern config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("[a-"),
				HostWritePath: ptr.To("/etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test not absolute hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("[a-z]"),
				HostWritePath: ptr.To("./etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test not absolute hostWritePath config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("[a-z]"),
				HostWritePath: ptr.To("etc/lvm"),
			},
			valid: false,
		},
		{
			desc: "test valid config",
			customData: &CsiDriverLvmConfig{
				DevicePattern: ptr.To("/dev/loop10[0,1]"),
				HostWritePath: ptr.To("/etc/lvm"),
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
