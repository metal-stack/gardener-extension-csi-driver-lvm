package healthcheck

import (
	"context"
	"time"

	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/csidriverlvm/v1alpha1"
	csidriverlvm "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/csi-driver-lvm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	defaultSyncPeriod = 60 * time.Second
	// DefaultAddOptions contains configuration for the health check controller.
	DefaultAddOptions = healthcheck.DefaultAddArgs{
		HealthCheckConfig: extensionsconfig.HealthCheckConfig{SyncPeriod: metav1.Duration{Duration: defaultSyncPeriod}},
	}
)

// RegisterHealthChecks registers health checks for each extension resource
// HealthChecks are grouped by extension (e.g worker), extension.type (e.g aws) and  Health Check Type (e.g SystemComponentsHealthy)
func RegisterHealthChecks(ctx context.Context, mgr manager.Manager, opts healthcheck.DefaultAddArgs) error {
	return healthcheck.DefaultRegistration(
		ctx,
		csidriverlvm.Type,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.ExtensionResource),
		func() client.ObjectList { return &extensionsv1alpha1.ExtensionList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.Extension{} },
		mgr,
		opts,
		nil,
		[]healthcheck.ConditionTypeToHealthCheck{
			{
				ConditionType: string(gardencorev1beta1.ShootSystemComponentsHealthy),
				HealthCheck:   general.CheckManagedResource(v1alpha1.SeedCsiDriverLvmResourceName),
			},
		},
		sets.Set[gardencorev1beta1.ConditionType]{},
	)
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return RegisterHealthChecks(ctx, mgr, DefaultAddOptions)
}
