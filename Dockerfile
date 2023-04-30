# syntax=docker/dockerfile:1

FROM golang:1.20-alpine3.16

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN apk add --no-cache upx || go version && go mod download
COPY . .
RUN CGO_ENABLED=0 go build -buildvcs=false -ldflags="-w -s" -o /itodo
RUN [ -e /usr/bin/upx ] && upx /itodo || echo
EXPOSE 80

ENTRYPOINT ["/itodo"]