FROM golang:alpine as build-env

RUN apk add git

COPY . /go/src/github.com/douban/ucloud-exporter
WORKDIR /go/src/github.com/douban/ucloud-exporter
# Build
ENV GOPATH=/go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -a -ldflags "-s -w" -o /go/bin/ucloud-exporter .

FROM library/alpine:3.15.0
COPY --from=build-env /go/bin/ucloud-exporter /usr/bin/ucloud-exporter
ENTRYPOINT ["ucloud-exporter"]
