FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /fetch-earnings ./cmd/jobs/fetch-earnings/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Tokyo
COPY --from=builder /fetch-earnings /fetch-earnings
CMD ["/fetch-earnings"]
