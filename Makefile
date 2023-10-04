BINARY_NAME=koherence
# TODO less bruteforce ?
KOHERENCE_FILES=$(shell find ${PWD} -type f -name "*.go")

# Control if static or dynamically linked (static by default)
export CGO_ENABLED:=0

.PHONY: build
build: ${BINARY_NAME}

${BINARY_NAME}: ${KOHERENCE_FILES}
	CGO_ENABLED=${CGO_ENABLED} go build -o ${BINARY_NAME} cmd/${BINARY_NAME}/*.go
	strip -x $@
