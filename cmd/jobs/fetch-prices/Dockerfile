FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /fetch-prices ./cmd/jobs/fetch-prices/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Tokyo
COPY --from=builder /fetch-prices /fetch-prices
CMD ["/fetch-prices"]
