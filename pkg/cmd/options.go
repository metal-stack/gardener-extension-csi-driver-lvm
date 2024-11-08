package cmd

import (
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"

	csidriverlvm "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/csi-driver-lvm"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/healthcheck"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the provider controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(csidriverlvm.ControllerName, csidriverlvm.AddToManager),
		controllercmd.Switch(extensionshealthcheckcontroller.ControllerName, healthcheck.AddToManager),
		controllercmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
