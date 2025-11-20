package csidriverlvm

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"

	gutil "github.com/gardener/gardener/extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/managedresources"

	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/go-logr/logr"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/config"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/apis/csidriverlvm/v1alpha1"
	"github.com/metal-stack/gardener-extension-csi-driver-lvm/pkg/imagevector"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	shootNamespace string = "kube-system"

	oldName        string = "csi-lvm"
	oldNamespace   string = "csi-lvm"
	oldProvisioner string = "metal-stack.io/csi-lvm"
)

// NewActuator returns an actuator responsible for Extension resources.
func NewActuator(mgr manager.Manager, config config.ControllerConfiguration) extension.Actuator {
	return &actuator{
		client:  mgr.GetClient(),
		decoder: serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
		config:  config,
	}
}

type actuator struct {
	client  client.Client
	decoder runtime.Decoder
	config  config.ControllerConfiguration
}

// Reconcile the Extension resource.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	csidriverlvmConfig := &v1alpha1.CsiDriverLvmConfig{}
	if ex.Spec.ProviderConfig != nil {
		_, _, err := a.decoder.Decode(ex.Spec.ProviderConfig.Raw, nil, csidriverlvmConfig)
		if err != nil {
			return fmt.Errorf("failed to decode provider config: %w", err)
		}
	}

	csidriverlvmConfig.ConfigureDefaults(a.config.DefaultHostWritePath, a.config.DefaultDevicePattern)
	if !csidriverlvmConfig.IsValid(log) {
		return fmt.Errorf("invalid csi-driver-lvm configuration")
	}

	isOldCsiLvmExisting, err := a.isOldCsiLvmExisting(ctx, ex.Namespace)
	if err != nil {
		return fmt.Errorf("failed to check if old csi-lvm is existing: %w", err)
	}
	if isOldCsiLvmExisting {
		return fmt.Errorf("assuming csi-lvm is still present due to existing storage class; csi-driver-lvm cannot run while csi-lvm is still deployed")
	}

	pluginObjects, err := a.getPluginObjects(csidriverlvmConfig)
	if err != nil {
		return fmt.Errorf("unable to get plugin objects: %w", err)
	}

	objects := []client.Object{}
	objects = append(objects, pluginObjects...)
	objects = append(objects, a.storageClasses(csidriverlvmConfig)...)

	shootResources, err := managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer).AddAllAndSerialize(objects...)
	if err != nil {
		return err
	}

	err = managedresources.CreateForShoot(ctx, a.client, ex.Namespace, v1alpha1.ShootCsiDriverLvmResourceName, "csi-driver-lvm-extension", false, shootResources)

	if err != nil {
		return err
	}

	log.Info("managed resource created successfully", "name", v1alpha1.ShootCsiDriverLvmResourceName)

	return nil
}

// Delete the Extension resource.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {

	log.Info("deleting managed resource")
	err := managedresources.Delete(ctx, a.client, ex.Namespace, v1alpha1.ShootCsiDriverLvmResourceName, false)

	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	err = managedresources.WaitUntilDeleted(timeoutCtx, a.client, ex.Namespace, v1alpha1.ShootCsiDriverLvmResourceName)
	if err != nil {
		return err
	}

	log.Info("successfully deleted managed resource")

	return nil
}

// ForceDelete the Extension resource
func (a *actuator) ForceDelete(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.Extension) error {
	return nil
}

// Restore the Extension resource.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, log, ex)
}

// Migrate the Extension resource.
func (a *actuator) Migrate(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return nil
}

func (a *actuator) storageClasses(csidriverlvmConfig *v1alpha1.CsiDriverLvmConfig) []client.Object {
	var (
		csidriverlvmLinearStorageClass = &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "csi-driver-lvm-linear",
			},
			Provisioner:          "lvm.csi.metal-stack.io",
			ReclaimPolicy:        ptr.To(corev1.PersistentVolumeReclaimDelete),
			VolumeBindingMode:    ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			AllowVolumeExpansion: pointer.Pointer(true),
			Parameters: map[string]string{
				"type": "linear",
			},
		}

		csidriverlvmMirrorStorageClass = &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "csi-driver-lvm-mirror",
			},
			Provisioner:          "lvm.csi.metal-stack.io",
			ReclaimPolicy:        ptr.To(corev1.PersistentVolumeReclaimDelete),
			VolumeBindingMode:    ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			AllowVolumeExpansion: pointer.Pointer(true),
			Parameters: map[string]string{
				"type": "mirror",
			},
		}

		csidriverlvmStripedStorageClass = &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "csi-driver-lvm-striped",
			},
			Provisioner:          "lvm.csi.metal-stack.io",
			ReclaimPolicy:        ptr.To(corev1.PersistentVolumeReclaimDelete),
			VolumeBindingMode:    ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			AllowVolumeExpansion: pointer.Pointer(true),
			Parameters: map[string]string{
				"type": "striped",
			},
		}

		csidriverlvmDefaultStorageClass = &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: oldName,
			},
			Provisioner:          "lvm.csi.metal-stack.io",
			ReclaimPolicy:        ptr.To(corev1.PersistentVolumeReclaimDelete),
			VolumeBindingMode:    ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			AllowVolumeExpansion: pointer.Pointer(true),
			Parameters: map[string]string{
				"type": "linear",
			},
		}

		storageClasses = []*storagev1.StorageClass{
			csidriverlvmDefaultStorageClass,
			csidriverlvmLinearStorageClass,
			csidriverlvmMirrorStorageClass,
			csidriverlvmStripedStorageClass,
		}

		objects []client.Object
	)

	// set default storageclass
	for _, sc := range storageClasses {
		if csidriverlvmConfig.DefaultStorageClass != nil && *csidriverlvmConfig.DefaultStorageClass == sc.Name {
			if sc.Annotations == nil {
				sc.Annotations = map[string]string{}
			}
			sc.Annotations["storageclass.kubernetes.io/is-default-class"] = "true"
		}
		objects = append(objects, sc)
	}
	return objects
}

func (a *actuator) getPluginObjects(csidriverlvmConfig *v1alpha1.CsiDriverLvmConfig) ([]client.Object, error) {
	csidriverlvmDriver := &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-driver-lvm",
			Namespace: shootNamespace,
		},
		Spec: storagev1.CSIDriverSpec{
			VolumeLifecycleModes: []storagev1.VolumeLifecycleMode{"Persistent", "Ephemeral"},
			PodInfoOnMount:       pointer.Pointer(true),
			AttachRequired:       pointer.Pointer(false),
			StorageCapacity:      pointer.Pointer(true),
		},
	}

	csidriverlvmServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-driver-lvm",
			Namespace: shootNamespace,
		},
	}

	csidriverlvmClusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-driver-lvm",
			Namespace: shootNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update", "patch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "update", "patch", "create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"csinodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"csistoragecapacities"},
				Verbs:     []string{"get", "list", "watch", "update", "patch", "create", "delete"},
			},
		},
	}

	csidriverlvmClusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-driver-lvm",
			Namespace: shootNamespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "csi-driver-lvm",
				Namespace: shootNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "csi-driver-lvm",
		},
	}

	csiNodeDriverRegistrarImage, err := imagevector.ImageVector().FindImage("csi-node-driver-registrar")
	if err != nil {
		return nil, fmt.Errorf("failed to find csi-node-driver-registrar image: %w", err)
	}

	livenessprobeImage, err := imagevector.ImageVector().FindImage("livenessprobe")
	if err != nil {
		return nil, fmt.Errorf("failed to find livenessprobe image: %w", err)
	}

	csiDriverLvmImage, err := imagevector.ImageVector().FindImage("csi-driver-lvm")
	if err != nil {
		return nil, fmt.Errorf("failed to find csi-driver-lvm image: %w", err)
	}

	csiAttacherImage, err := imagevector.ImageVector().FindImage("csi-attacher")
	if err != nil {
		return nil, fmt.Errorf("failed to find csi-attacher image: %w", err)
	}

	csiResizerImage, err := imagevector.ImageVector().FindImage("csi-resizer")
	if err != nil {
		return nil, fmt.Errorf("failed to find csi-resizer image: %w", err)
	}

	csiDriverLvmControllerImage, err := imagevector.ImageVector().FindImage("csi-driver-lvm-controller")
	if err != nil {
		return nil, fmt.Errorf("failed to find eviction-controller image: %w", err)
	}

	csiProvisionerImage, err := imagevector.ImageVector().FindImage("csi-provisioner")
	if err != nil {
		return nil, fmt.Errorf("failed to find csi-provisioner image: %w", err)
	}

	var terminationPolicy = corev1.TerminationMessageReadFile
	var mountPropagation = corev1.MountPropagationBidirectional

	var hostPathTypeCreate = corev1.HostPathDirectoryOrCreate
	var hostPathTypeDir = corev1.HostPathDirectory

	csidriverlvmDaemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-driver-lvm",
			Namespace: shootNamespace,
		},
		Spec: appsv1.DaemonSetSpec{
			RevisionHistoryLimit: ptr.To(int32(10)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "csi-driver-lvm",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "csi-driver-lvm",
					},
				}, Spec: corev1.PodSpec{
					ServiceAccountName: "csi-driver-lvm",
					Containers: []corev1.Container{
						// controller
						{
							Name:            "csi-attacher",
							Image:           csiAttacherImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args:            []string{"--v=5", "--csi-address=/csi/csi.sock"},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(true),
								Privileged:             pointer.Pointer(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
							},
						},
						{
							Name:            "csi-provisioner",
							Image:           csiProvisionerImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args: []string{
								"--v=5", "--csi-address=/csi/csi.sock",
								"--feature-gates=Topology=true",
								"--enable-capacity",
								"--node-deployment",
								"--strict-topology",
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(true),
								Privileged:             pointer.Pointer(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
							},
						},
						{
							Name:            "csi-resizer",
							Image:           csiResizerImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args:            []string{"--v=5", "--csi-address=/csi/csi.sock"},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(true),
								Privileged:             pointer.Pointer(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
							},
						},
						{
							Name:            "csi-driver-lvm-controller",
							Image:           csiDriverLvmControllerImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args:            []string{"--leader-elect", "--health-probe-bind-address=:8081"},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Pointer(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8081),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8081),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
						// plugin
						{
							Name:            "csi-node-driver-registrar",
							Image:           csiNodeDriverRegistrarImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args:            []string{"--v=5", "--csi-address=/csi/csi.sock", "--kubelet-registration-path=/var/lib/kubelet/plugins/csi-driver-lvm/csi.sock"},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(false),
								Privileged:             pointer.Pointer(true),
							},
							Env: []corev1.EnvVar{
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: terminationPolicy,
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
								{MountPath: "/var/lib/kubelet/plugins/csi-driver-lvm/csi.sock", Name: "socket-dir"},
								{MountPath: "/registration", Name: "registration-dir"},
							},
						},
						{
							Name:            "csi-driver-lvm-plugin",
							Image:           csiDriverLvmImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args: []string{
								"--drivername=lvm.csi.metal-stack.io",
								"--endpoint=/csi/csi.sock",
								"--hostwritepath=" + pointer.SafeDeref(csidriverlvmConfig.HostWritePath),
								"--devices=" + pointer.SafeDeref(csidriverlvmConfig.DevicePattern),
								"--nodeid=$(KUBE_NODE_NAME)",
								"--vgname=csi-lvm",
							},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(false),
								Privileged:             pointer.Pointer(true),
							},
							Env: []corev1.EnvVar{
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								FailureThreshold:    5,
								InitialDelaySeconds: 10,
								PeriodSeconds:       2,
								SuccessThreshold:    1,
								TimeoutSeconds:      3,
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(9898),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							Ports: []corev1.ContainerPort{{
								Name:          "healthz",
								Protocol:      corev1.ProtocolTCP,
								ContainerPort: 9898,
							}},
							TerminationMessagePath:   "/tmp/termination-log", // not mounting to /dev since it is mounted as hostpath
							TerminationMessagePolicy: terminationPolicy,
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
								{MountPath: "/var/lib/kubelet/pods", Name: "mountpoint-dir", MountPropagation: &mountPropagation},
								{MountPath: "/var/lib/kubelet/plugins", Name: "plugins-dir", MountPropagation: &mountPropagation},
								{MountPath: "/dev", Name: "dev-dir", MountPropagation: &mountPropagation},
								{MountPath: "/lib/modules", Name: "mod-dir"},
								{MountPath: "/etc/lvm/backup", Name: "lvmbackup", MountPropagation: &mountPropagation},
								{MountPath: "/etc/lvm/cache", Name: "lvmcache", MountPropagation: &mountPropagation},
								{MountPath: "/etc/lvm/archive", Name: "lvmarchive", MountPropagation: &mountPropagation},
								{MountPath: "/etc/lvm/lock", Name: "lvmlock", MountPropagation: &mountPropagation},
							},
						},
						{
							Name:            "livenessprobe",
							Image:           livenessprobeImage.String(),
							ImagePullPolicy: *csidriverlvmConfig.PullPolicy,
							Args: []string{
								"--csi-address=/csi/csi.sock",
								"--health-port=9898",
							},
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: pointer.Pointer(true),
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: terminationPolicy,
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/csi", Name: "socket-dir"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "socket-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins/csi-driver-lvm",
									Type: &hostPathTypeCreate,
								},
							},
						},
						{
							Name: "mountpoint-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/pods",
									Type: &hostPathTypeCreate,
								},
							},
						},
						{
							Name: "registration-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins_registry",
									Type: &hostPathTypeDir,
								},
							},
						},
						{
							Name: "plugins-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins",
									Type: &hostPathTypeDir,
								},
							},
						},
						{
							Name: "dev-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev",
									Type: &hostPathTypeDir,
								},
							},
						},
						{
							Name: "mod-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
						{
							Name: "lvmcache",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pointer.SafeDeref(csidriverlvmConfig.HostWritePath) + "/cache",
									Type: &hostPathTypeCreate,
								},
							},
						},
						{
							Name: "lvmarchive",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pointer.SafeDeref(csidriverlvmConfig.HostWritePath) + "/archive",
									Type: &hostPathTypeCreate,
								},
							},
						},
						{
							Name: "lvmbackup",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pointer.SafeDeref(csidriverlvmConfig.HostWritePath) + "/backup",
									Type: &hostPathTypeCreate,
								},
							},
						},
						{
							Name: "lvmlock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: pointer.SafeDeref(csidriverlvmConfig.HostWritePath) + "/lock",
									Type: &hostPathTypeCreate,
								},
							},
						},
					},
				},
			},
		},
	}

	objects := []client.Object{
		csidriverlvmDriver,
		csidriverlvmServiceAccount,
		csidriverlvmClusterRole,
		csidriverlvmClusterRoleBinding,
		csidriverlvmDaemonSet,
	}

	return objects, nil
}

func (a *actuator) isOldCsiLvmExisting(ctx context.Context, shootNamespace string) (bool, error) {
	_, shootClient, err := gutil.NewClientForShoot(ctx, a.client, shootNamespace, client.Options{}, extensionsconfig.RESTOptions{})

	if err != nil {
		return true, fmt.Errorf("failed to create shoot client: %w", err)
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: oldNamespace,
		},
	}
	err = shootClient.Get(ctx, client.ObjectKeyFromObject(namespace), namespace)

	if err == nil {
		return true, nil
	} else if !apierrors.IsNotFound(err) {
		return true, fmt.Errorf("error while getting old csi-lvm namespace: %w", err)
	}

	storageClassList := &storagev1.StorageClassList{}
	err = shootClient.List(ctx, storageClassList)
	if err != nil {
		return false, fmt.Errorf("failed to list storage classes: %w", err)
	}

	for _, sc := range storageClassList.Items {
		if sc.Provisioner == oldProvisioner {
			return true, nil
		}
	}
	return false, nil
}
