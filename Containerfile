FROM golang:alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG JQ_VERSION=1.7

RUN apk update && apk add --no-cache bash=5.2.15-r5 git=2.40.1-r0 make=4.4.1-r1 binutils=2.40-r7 wget=1.21.4-r0 \
	&& wget --progress=dot:giga "https://github.com/jqlang/jq/releases/download/jq-${JQ_VERSION}/jq-${GOOS}-${GOARCH}" -O /usr/bin/jq \
	&& chmod +x /usr/bin/jq

WORKDIR $GOPATH/src/github.com/wiremind/koherence

COPY . .

RUN make koherence \
	&& mv koherence /usr/bin/

FROM scratch as runtime

COPY --from=build /usr/bin/koherence /koherence

COPY --from=build /usr/bin/jq /jq

ENTRYPOINT ["/koherence"]
