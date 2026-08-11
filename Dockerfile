FROM registry.sfbp.cn/hpc-img-pub/oven/bun:1.3.14 AS builder

WORKDIR /build
COPY web/default/package.json .
COPY web/default/bun.lock .
RUN bun install --frozen-lockfile
COPY ./web/default .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

FROM registry.sfbp.cn/hpc-img-pub/oven/bun:1.3.14 AS builder-classic

WORKDIR /build
COPY web/classic/package.json .
COPY web/classic/bun.lock .
RUN bun install --frozen-lockfile
COPY ./web/classic .
COPY ./VERSION .
RUN VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

FROM registry.sfbp.cn/hpc-img-pub/golang:1.26.1-alpine AS builder2
ENV GO111MODULE=on CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}
ENV GOEXPERIMENT=greenteagc

WORKDIR /build

ADD go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/default/dist
COPY --from=builder-classic /build/dist ./web/classic/dist
RUN go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

FROM registry.sfbp.cn/hpc-img-pub/debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata libasan8 wget \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

COPY --from=builder2 /build/new-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/new-api"]
