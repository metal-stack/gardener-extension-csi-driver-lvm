package cmd

import (
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"

	csidriverlvm "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/csi-driver-lvm"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/healthcheck"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/kapiserver"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the provider controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(csidriverlvm.ControllerName, csidriverlvm.AddToManager),
		controllercmd.Switch(extensionshealthcheckcontroller.ControllerName, healthcheck.AddToManager),
		controllercmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}

// WebhookSwitchOptions are the webhookcmd.SwitchOptions for the provider webhooks.
func WebhookSwitchOptions() *webhookcmd.SwitchOptions {
	return webhookcmd.NewSwitchOptions(
		webhookcmd.Switch("csi-driver-lvm-webhook", kapiserver.New),
	)
}
