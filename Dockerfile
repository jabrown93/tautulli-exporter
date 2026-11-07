# Static CGO_ENABLED=0 cross-compile on the DHI Go toolchain, scratch runtime,
# nonroot.
FROM --platform=$BUILDPLATFORM dhi.io/golang:1.27.3-alpine-dev@sha256:bc66efe51594cd50a6cf42d36f8521c6d643bd9be206acf606753ed15b23f0e1 AS builder

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
