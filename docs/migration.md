# Migration from csi-lvm to csi-driver-lvm

The migration from the deprecated [`csi-lvm`](https://github.com/metal-stack/csi-lvm) to the new `csi-driver-lvm` must be performed in multiple steps. In here we assume you are currently using the [gardener-extension-provider-metal](https://github.com/metal-stack/gardener-extension-provider-metal).

1. Install `gardener-extension-csi-driver-lvm`. As long as the `csi-lvm` is still present in the shoot, this will not yet install the `csi-driver-lvm`.
2. Disable `csi-lvm` in `gardener-extension-provider-metal` by setting `featureGates.disableCsiLvm` to `true`. This allows the `csi-driver-lvm` to be installed.
3. Wait until the shoot has been successfully reconciled.
4. Make sure the `csi-driver-lvm` has been installed and that the `csi-lvm` storage class has been created.
5. Make sure to roll all nodes. Otherwise volumes will stop working after the next restart of a machine.

## Issues

### Why is the Drop-in replacement not possible?

Deploying the new csi-driver-lvm with the same provisioner-name as the old one is not possible, as it causes errors when using k8s sidecar images for controllers.

The provisioner name contains "/", which causes problems with node registrar directories (**metal-stack.io/csi-lvm**):

```sh
I1015 08:09:23.292306 1 node_register.go:53] Starting Registration Server at: /registration/metal-stack.io/csi-lvm-reg.sock

E1015 08:09:23.292482 1 node_register.go:56] failed to listen on socket: /registration/metal-stack.io/csi-lvm-reg.sock with error: listen unix /registration/metal-stack.io/csi-lvm-reg.sock: bind: no such file or directory
```

This problem requires a more complex migration.

## Manual migration
The migration solution so far has been tested manually:

1. create old controller & provisioner
2. create pvcs & pod
3. write files to volumes
4. delete old controller & provisioner
5. install new controller & provisioner with helm
6. add additional storage class with name `csi-lvm` and type linear
    1. mimics old storage class
    2. default storage class (not supported yet -> see default storage class of `gardener-extension-provider-metal`)
7. create new pvcs
8. create new pod with old and new pvcs and test
