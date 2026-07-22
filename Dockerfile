FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener ./cmd/url-shortener

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /url-shortener ./url-shortener
COPY config ./config

EXPOSE 8082

ENTRYPOINT ["./url-shortener"]
