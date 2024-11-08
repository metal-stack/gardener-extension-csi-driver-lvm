package app

import (
	"context"
	"fmt"
	"os"

	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/csidriverlvm/install"

	corev1 "k8s.io/api/core/v1"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	heartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	csidriverlvmcmd "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/cmd"
	controller "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/csi-driver-lvm"

	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	ghealth "github.com/gardener/gardener/pkg/healthz"
	componentbaseconfig "k8s.io/component-base/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// var log = runtimelog.Log.WithName("gardener-extension-csi-driver-lvm")

const ExtensionName = "extension-csi-driver-lvm"

type Options struct {
	generalOptions      *controllercmd.GeneralOptions
	csidriverlvmOptions *csidriverlvmcmd.AuthOptions
	restOptions         *controllercmd.RESTOptions
	managerOptions      *controllercmd.ManagerOptions
	controllerOptions   *controllercmd.ControllerOptions
	heartbeatOptions    *heartbeatcmd.Options
	healthOptions       *controllercmd.ControllerOptions
	controllerSwitches  *controllercmd.SwitchOptions
	reconcileOptions    *controllercmd.ReconcilerOptions
	optionAggregator    controllercmd.OptionAggregator
}

func NewOptions() *Options {
	options := &Options{
		generalOptions:      &controllercmd.GeneralOptions{},
		csidriverlvmOptions: &csidriverlvmcmd.AuthOptions{},
		restOptions:         &controllercmd.RESTOptions{},
		managerOptions: &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(ExtensionName),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			MetricsBindAddress:      ":8080",
			HealthBindAddress:       ":8081",
		},

		// options for the controlplane controller
		controllerOptions: &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		},

		heartbeatOptions: &heartbeatcmd.Options{
			// This is a default value.
			ExtensionName:        ExtensionName,
			RenewIntervalSeconds: 30,
			Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		healthOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		controllerSwitches: csidriverlvmcmd.ControllerSwitchOptions(),
		reconcileOptions:   &controllercmd.ReconcilerOptions{},
	}

	options.optionAggregator = controllercmd.NewOptionAggregator(
		options.generalOptions,
		options.csidriverlvmOptions,
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		controllercmd.PrefixOption("heartbeat-", options.heartbeatOptions),
		controllercmd.PrefixOption("healthcheck-", options.healthOptions),
		options.controllerSwitches,
		options.reconcileOptions,
	)

	return options
}

func (options *Options) run(ctx context.Context) error {
	log.Info("starting " + ExtensionName)

	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfig.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, options.restOptions.Completed().Config)

	log.Info("applied rest config")

	mgrOpts := options.managerOptions.Completed().Options()

	log.Info("completed mgr-options")

	mgrOpts.Client = client.Options{
		Cache: &client.CacheOptions{
			DisableFor: []client.Object{
				&corev1.Secret{},
				&corev1.ConfigMap{},
			},
		},
	}

	mgr, err := manager.New(options.restOptions.Completed().Config, mgrOpts)
	if err != nil {
		return fmt.Errorf("could not instantiate controller-manager: %w", err)
	}

	log.Info("completed rest-options")

	err = extensionscontroller.AddToScheme(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("could not add mgr-scheme to extension-controller: %w", err)
	}

	log.Info("added mgr-scheme to extensionscontroller")

	err = install.AddToScheme(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("could not add mgr-scheme to installation")
	}

	log.Info("added mgr-schme to installation")

	ctrlConfig := options.csidriverlvmOptions.Completed()
	ctrlConfig.Apply(&controller.DefaultAddOptions.Config)

	options.controllerOptions.Completed().Apply(&controller.DefaultAddOptions.ControllerOptions)
	options.reconcileOptions.Completed().Apply(&controller.DefaultAddOptions.IgnoreOperationAnnotation)
	options.heartbeatOptions.Completed().Apply(&heartbeatcontroller.DefaultAddOptions)

	if err := options.controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
		return fmt.Errorf("could not add controllers to manager: %w", err)
	}

	if err := mgr.AddReadyzCheck("informer-sync", ghealth.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
		return fmt.Errorf("could not add ready check for informers: %w", err)
	}

	log.Info("added readyzcheck")

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("could not add health check to manager: %w", err)
	}

	log.Info("added healthzcheck")

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}

	log.Info("started successfully")

	return nil
}
