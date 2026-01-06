# Build stage for s3fs-fuse from source
FROM alpine:3.17 as s3fsbuild

# 设置环境变量
ENV http_proxy=http://10.249.1.2:8118
ENV https_proxy=http://10.249.1.2:8118


RUN apk add --no-cache \
    alpine-sdk automake make fuse3-dev fuse3 \
    libxml2-dev openssl-dev curl-dev \
    ca-certificates git mailcap pkgconf libstdc++ libgcc 

RUN cd /build && \
    git clone https://github.com/Sun1Plus/s3fs-fuse.git && \
    git switch custom-credentials && \
    cd s3fs-fuse && \
    mkdir -p m4 && \
    ./autogen.sh && \
    ./configure --with-openssl && \
    make && \
    make install

# Build stage for Go CSI driver
FROM golang:1.19-alpine as gobuild

WORKDIR /build
ADD go.mod go.sum /build/
RUN go mod download -x
ADD cmd /build/cmd
ADD pkg /build/pkg
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./s3driver ./cmd/s3driver

# Final runtime image
FROM alpine:3.17
LABEL maintainers="Vitaliy Filippov <vitalif@yourcmc.ru>"
LABEL description="csi-s3 slim image"

RUN apk add --no-cache \
    fuse3 fuse3-libs libxml2 openssl curl mailcap rclone libstdc++

# COPY --from=s3fsbuild /usr/local/bin/s3fs /usr/bin/s3fs
# ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /usr/bin/geesefs
# RUN chmod 755 /usr/bin/geesefs

COPY --from=gobuild /build/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
