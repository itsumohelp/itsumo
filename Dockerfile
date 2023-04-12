# syntax=docker/dockerfile:1

FROM golang:1.20-alpine3.16

WORKDIR /app

COPY . /app
RUN go mod download
RUN go build -o /itodo

EXPOSE 80

CMD ["/itodo"]