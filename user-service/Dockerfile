FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/user-service ./app

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/user-service /usr/local/bin/user-service
COPY migration ./migration
COPY config.dev.yaml ./config.dev.yaml

EXPOSE 7705

CMD ["user-service", "-c", "config.dev.yaml"]
