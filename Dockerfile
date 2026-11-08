# Static CGO_ENABLED=0 cross-compile on the DHI Go toolchain, scratch runtime,
# nonroot.
FROM --platform=$BUILDPLATFORM dhi.io/golang:1.27.3-alpine-dev@sha256:95894969dcf353c87cd1b9730e599e1d3c3b7519c8966d28120859609269df4d AS builder

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
