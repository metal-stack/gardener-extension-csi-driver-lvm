# Migration from csi-lvm to csi-driver-lvm

## Objective

The goal of this document is to provide instructions for migrating from the old [csi-lvm](https://github.com/metal-stack/csi-lvm/tree/master) to the new [csi-driver-lvm](https://github.com/metal-stack/csi-driver-lvm).

## Issues

### Drop-in replacement not possible

Deploying the new csi-driver-lvm with the same provisioner-name as the old one is not possible, as it causes errors when using k8s sidecar images for controllers.

The provisioner name contains "/", which causes problems with node registrar directories (**metal-stack.io/csi-lvm**):

```sh
I1015 08:09:23.292306 1 node_register.go:53] Starting Registration Server at: /registration/metal-stack.io/csi-lvm-reg.sock

E1015 08:09:23.292482 1 node_register.go:56] failed to listen on socket: /registration/metal-stack.io/csi-lvm-reg.sock with error: listen unix /registration/metal-stack.io/csi-lvm-reg.sock: bind: no such file or directory
```

This problem requires a more complex migration.

## Solution

### Local motivation
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

### Migration

To achieve this behaviour for csi-lvm, provided by [gardener-extension-provider-metal](https://github.com/metal-stack/gardener-extension-provider-metal/tree/master), we need to add the following workflow:

1. Add a feature gate to `gardener-extension-provider-metal` to disable csi-lvm.
2. When deploying `gardener-extension-csi-driver-lvm`, stop reconciliation if old provisioner is still available.
