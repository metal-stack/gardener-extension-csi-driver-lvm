package kapiserver

import (
	"context"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/csidriverlvm/v1alpha1"
	csidriverlvm "github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/controller/csi-driver-lvm"

	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(logger logr.Logger, mgr manager.Manager) genericmutator.Ensurer {
	return &ensurer{
		client:  mgr.GetClient(),
		decoder: serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
		logger:  logger.WithName("csi-driver-lvm-controlplane-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client  client.Client
	decoder runtime.Decoder
	logger  logr.Logger
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, _ *appsv1.Deployment) error {
	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return err
	}

	if cluster.Shoot.DeletionTimestamp != nil && !cluster.Shoot.DeletionTimestamp.IsZero() {
		e.logger.Info("skip mutating api server because shoot is in deletion")
		return nil
	}

	namespace := cluster.ObjectMeta.Name

	ex := &extensionsv1alpha1.Extension{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csidriverlvm.Type,
			Namespace: namespace,
		},
	}
	err = e.client.Get(ctx, client.ObjectKeyFromObject(ex), ex)
	if err != nil {
		return fmt.Errorf("unable to find extension resource. this extension needs to be configured with lifecycle policy BeforeKubeAPIServer")
	}

	csiDriverLvm := &v1alpha1.CsiDriverLvmConfig{}
	if ex.Spec.ProviderConfig != nil {
		if _, _, err := e.decoder.Decode(ex.Spec.ProviderConfig.Raw, nil, csiDriverLvm); err != nil {
			return fmt.Errorf("failed to decode provider config: %w", err)
		}
	}

	template := &new.Spec.Template
	ps := &template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		e.logger.Info("ensuring kube-apiserver deployment")
		ensureKubeAPIServerCommandLineArgs(c)
		ensureVolumeMounts(c)
		ensureVolumes(ps)
	}

	// ? whats this?
	// template.Labels["networking.resources.gardener.cloud/to-audit-webhook-backend-tcp-9880"] = "allowed"

	return nil
}

func ensureVolumeMounts(c *corev1.Container) {
	// c.VolumeMounts = extensionswebhook.EnsureVolumeMountWithName(c.VolumeMounts, corev1.VolumeMount{
	// 	Name:      "audit-webhook-config",
	// 	ReadOnly:  true,
	// 	MountPath: "/etc/audit-webhook/config",
	// })
}

func ensureVolumes(ps *corev1.PodSpec) {
	// ps.Volumes = extensionswebhook.EnsureVolumeWithName(ps.Volumes, corev1.Volume{
	// 	Name: "audit-webhook-config",
	// 	VolumeSource: corev1.VolumeSource{
	// 		Secret: &corev1.SecretVolumeSource{
	// 			SecretName: "audit-webhook-config",
	// 		},
	// 	},
	// })
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) {
	// c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--audit-webhook-config-file=", "/etc/audit-webhook/config/audit-webhook-config.yaml")
}
