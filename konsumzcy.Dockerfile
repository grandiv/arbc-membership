# KonsumZcy (member store) build.
# Run scripts/stage.sh first to populate .kzcy/ (libs + engine source). Build
# context = this repo root; see docker-compose.yml.
# --platform=$BUILDPLATFORM: the builder runs natively (e.g. arm64 on the dev
# Mac) and CROSS-compiles to $TARGETARCH, so we can build amd64 images locally
# and ship them — the reduced-spec VPS no longer compiles anything.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev

COPY .kzcy/libs/ /libs/
COPY .kzcy/services/KonsumZcy/KonsumZcy/ ./

# Rewrite the relative replace paths (../../../libs/) to the in-image location.
RUN sed -i 's|../../../libs/|/libs/|g' go.mod

# izcy-engine VPS blocks Google's module proxy + checksum DB.
ENV GOPROXY=https://goproxy.io,https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct
ENV GOSUMDB=off
ARG TARGETOS TARGETARCH
RUN go mod download && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o konsumzcy ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/konsumzcy .
ENV GIN_MODE=release
EXPOSE 8084
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8084/health || exit 1
CMD ["./konsumzcy"]
