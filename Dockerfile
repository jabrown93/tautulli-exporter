# Static CGO_ENABLED=0 cross-compile on the DHI Go toolchain, scratch runtime,
# nonroot.
FROM --platform=$BUILDPLATFORM dhi.io/golang:1.26.6-dev@sha256:e7f194366f8993d4c33eeeb24d96eee1b615c695127606917e93a9a81c3bd8ea AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags='-w -s' -o /out/tautulli_exporter .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/tautulli_exporter /tautulli_exporter

USER 65532:65532
EXPOSE 9487
ENTRYPOINT ["/tautulli_exporter"]
