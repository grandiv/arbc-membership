# PromoZcy (voucher / campaign engine) build.
# Run scripts/stage.sh first to populate .kzcy/. See docker-compose.yml.
# --platform=$BUILDPLATFORM: builder runs natively, cross-compiles to $TARGETARCH
# (see konsumzcy.Dockerfile) so amd64 images can be built locally and shipped.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev

COPY .kzcy/libs/ /libs/
COPY .kzcy/services/PromoZcy/PromoZcy/ ./

RUN sed -i 's|../../../libs/|/libs/|g' go.mod

ENV GOPROXY=https://goproxy.io,https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct
ENV GOSUMDB=off
ARG TARGETOS TARGETARCH
RUN go mod download && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o promozcy ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/promozcy .
ENV GIN_MODE=release
EXPOSE 8082
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8082/health || exit 1
CMD ["./promozcy"]
