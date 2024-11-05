//go:generate sh -c "bash $GARDENER_HACK_DIR/generate-controller-registration.sh csi-driver-lvm . $(cat ../../VERSION) ../../example/controller-registration.yaml Extension:csi-driver-lvm"

// Package chart enables go:generate support for generating the correct controller registration.
package chart
