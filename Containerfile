FROM golang:alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG JQ_VERSION=1.7

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache bash git make binutils wget \
	&& wget --progress=dot:giga "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-${GOOS}-${GOARCH}" -O /usr/bin/jq \
	&& chmod +x /usr/bin/jq

WORKDIR $GOPATH/src/github.com/wiremind/koherence

COPY . .

RUN make koherence && mv koherence /usr/bin/


FROM busybox:stable as runtime

COPY --from=build /usr/bin/koherence /usr/bin/koherence
COPY --from=build /usr/bin/jq /usr/bin/jq

ENTRYPOINT ["/usr/bin/koherence"]
