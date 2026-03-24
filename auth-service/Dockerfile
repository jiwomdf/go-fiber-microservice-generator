FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service ./app

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/auth-service /usr/local/bin/auth-service
COPY migration ./migration
COPY config.dev.yaml ./config.dev.yaml

EXPOSE 7704

CMD ["auth-service", "-c", "config.dev.yaml"]
