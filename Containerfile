FROM golang:alpine

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR $GOPATH/src/github.com/wiremind/koherence
COPY . .

RUN apk update && apk add --virtual .build-deps bash git make binutils \
	&& make koherence \
	&& mv koherence /usr/bin/ \
	&& rm -rf $GOPATH \
	&& apk del .build-deps \
	&& rm -rf /var/cache/apk

FROM scratch as runtime
COPY --from=0 /usr/bin/koherence /koherence
ENTRYPOINT ["/koherence"]
