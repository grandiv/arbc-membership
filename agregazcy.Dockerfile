# AgregaZcy (analytics timeline + forecasting) build.
# Run scripts/stage.sh first to populate .kzcy/. See docker-compose.yml.
# --platform=$BUILDPLATFORM: builder runs natively, cross-compiles to $TARGETARCH
# (see konsumzcy.Dockerfile) so amd64 images can be built locally and shipped.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev

COPY .kzcy/libs/ /libs/
COPY .kzcy/services/AgregaZcy/AgregaZcy-BI-Go/ /bi-go/
COPY .kzcy/services/AgregaZcy/AgregaZcy/ ./

RUN sed -i 's|../../../libs/|/libs/|g; s|../AgregaZcy-BI-Go|/bi-go|g' go.mod

ENV GOPROXY=https://goproxy.io,https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct
ENV GOSUMDB=off
ARG TARGETOS TARGETARCH
RUN go mod download && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o agregazcy ./cmd/server/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/agregazcy .
ENV GIN_MODE=release
EXPOSE 5900
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:5900/health || exit 1
CMD ["./agregazcy"]
