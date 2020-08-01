FROM golang:1.14-alpine

WORKDIR /go/src/apparea/
ADD go.mod go.sum ./
ADD server/ server/

WORKDIR /go/src/apparea/server/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine

RUN apk add openssh-client

COPY --from=0 /go/src/apparea/server/server /usr/local/bin/apparea
ADD ./server/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT [ "docker-entrypoint.sh" ]
