FROM golang:alpine

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG JQ_VERSION=1.7

WORKDIR $GOPATH/src/github.com/wiremind/koherence
COPY . .

RUN apk update && apk add --virtual .build-deps bash git make binutils wget \
	&& make koherence \
	&& mv koherence /usr/bin/ \
	&& rm -rf $GOPATH \
	&& apk del .build-deps \
	&& rm -rf /var/cache/apk \
	&& wget "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-${GOOS}-${GOARCH}" -O /usr/bin/jq \
	&& chmod +x /usr/bin/jq

FROM scratch as runtime
COPY --from=0 /usr/bin/koherence /koherence
COPY --from=0 /usr/bin/jq /jq
ENTRYPOINT ["/koherence"]
