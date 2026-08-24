# Static CGO_ENABLED=0 cross-compile on the DHI Go toolchain, scratch runtime,
# nonroot.
FROM --platform=$BUILDPLATFORM dhi.io/golang:1.27.0-alpine-dev@sha256:9558afe9b05f8d8429980a9e06d365120c2510354ec5168f51e4602bb9a4407c AS builder

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
