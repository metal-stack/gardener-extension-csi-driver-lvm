FROM golang:1.23 AS builder

WORKDIR /go/src/github.com/metal-stack/gardener-extension-csi-driver-lvm
COPY . .
RUN make install \
 && strip /go/bin/gardener-extension-csi-driver-lvm

FROM alpine:3.21
WORKDIR /
COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-csi-driver-lvm /gardener-extension-csi-driver-lvm
CMD ["/gardener-extension-csi-driver-lvm"]
