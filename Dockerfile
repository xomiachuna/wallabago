# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder

# download and extract /tools/upx for executable compression
WORKDIR /tools
RUN apt-get update && apt-get install xz-utils --no-install-recommends -yqq && \
    UPX_VERSION="5.0.1" && \
        wget -O upx.tar.gz \
            "https://github.com/upx/upx/releases/download/v$UPX_VERSION/upx-$UPX_VERSION-amd64_linux.tar.xz" && \
    tar xvf upx.tar.gz --strip-components=1

WORKDIR /app
# copy dependency info
COPY go.mod go.sum /app/
# pre-fill module cache (with more verbosity)
RUN go mod download -x
# copy the rest
COPY . /app/

# build an executable with race detection and embed vcs info
# and then compress it with upx
RUN CGOENABLED=1 GOOS=linux \
    go build -C cmd/wallabago-api \
        -o /app/server \
        -race \
        -buildvcs=true && \
    /tools/upx -v --best /app/server

# run tests
FROM builder AS tester
RUN go test -v ./...

# use smallest image necessary as described in 
# https://github.com/GoogleContainerTools/distroless/tree/main/base
# static-* contains no libssl and no glibc, root user, /tmp and tzdata 
# (incompatible with CGOENABLED=1)
# base-nossl-* contains no libssl but contains glibc and everything from static-*
# base-* contains libssl and glibc and everything from static-*
FROM gcr.io/distroless/base-nossl-debian12:nonroot AS runtime
USER nonroot:nonroot
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080
CMD [ "/app/server" ]
