#!/bin/sh
# lint.sh - Script to run the custom linters
LINTERS_PKG=./pkg/linters

#---------------------------errchecklog---------------------------
ERRCHECKLOG="${LINTERS_PKG}/errchecklog/cmd/main.go"
IFACE_PKG="go.avito.ru/av/service-ratings-users-composition/internal/pkg"
IFACE_NAME="Log"
TARGET="./internal/api/handler/* ./internal/rpc/*"
go build -o bin/errchecklog ${ERRCHECKLOG}
./bin/errchecklog -ifacepkg=${IFACE_PKG} -ifacename=${IFACE_NAME} ${TARGET}
