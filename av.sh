#!/usr/bin/env bash
cd
pwd
ls -la
echo "aa"
ls -la /go/src/github.com/half2me/antgo
export PATH="/mingw64/bin:${PATH}" && \
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig:${PKG_CONFIG_PATH}" && \
export GOROOT=/mingw64/lib/go && \
export GOPATH=/go && \
export CGO_ENABLED=1 && \
pacman --noconfirm -S \
    mingw64/mingw-w64-x86_64-go \
    mingw64/mingw-w64-x86_64-libusb && \
go version
go get github.com/half2me/antgo/... && \
go build -o antgo-win64.exe -i github.com/half2me/antgo && \
echo "Build complete!"
ls -la