# gardener-extension-csi-driver-lvm

Provides a Gardener extension for managing [csi-driver-lvm](https://github.com/metal-stack/csi-driver-lvm) for a shoot cluster.

The extension checks for the old [csi-lvm](https://github.com/metal-stack/csi-lvm/tree/master) and stops reconciling if the old driver is stil available.
If not the extension will reconcile the new `csi-driver-lvm`.

## Development

This extension can be developed in the gardener-local devel environment. Before make sure you have created loop-devices on your machine.

```sh
for i in 100 101; do fallocate -l 1G loop${i}.img ; sudo losetup /dev/loop${i} loop${i}.img; done
sudo losetup -a
# use this for recreation or cleanup
# for i in 100 101; do sudo losetup -d /dev/loop${i}; rm -f loop${i}.img; done
```

1. Start up the local devel environment
1. The extension's docker image can be pushed into Kind using `make push-to-gardener-local`
1. Install the extension `kubectl apply -k example/`
1. Parametrize the `example/shoot.yaml` and apply with `kubectl -f example/shoot.yaml`

