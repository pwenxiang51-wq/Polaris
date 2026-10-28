FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -p=1 -trimpath -ldflags="-s -w" -o /out/polaris ./cmd/polaris

FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 polaris \
    && adduser -S -D -H -u 10001 -G polaris polaris \
    && mkdir -p /data \
    && chown polaris:polaris /data
COPY --from=build /out/polaris /usr/local/bin/polaris
USER polaris:polaris
ENV POLARIS_LISTEN=0.0.0.0:8080 POLARIS_DATA_DIR=/data
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/polaris"]
