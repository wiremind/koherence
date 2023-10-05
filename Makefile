GO?=go

BINARY_NAME=koherence
# TODO less bruteforce ?
KOHERENCE_FILES=$(shell find ${PWD} -type f -name "*.go")

VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty='.m' --always)
REVISION=$(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)
PKG=github.com/wiremind/koherence

# Control if static or dynamically linked (static by default)
export CGO_ENABLED:=0

GO_GCFLAGS?=
GO_LDFLAGS=-ldflags '-X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.Revision=$(REVISION) -X $(PKG)/version.Package=$(PKG)'

.PHONY: build
build: ${BINARY_NAME}

${BINARY_NAME}: ${KOHERENCE_FILES}
	${GO} build ${GO_GCFLAGS} ${GO_LDFLAGS} -o $@ cmd/${BINARY_NAME}/*.go
	strip -x $@
