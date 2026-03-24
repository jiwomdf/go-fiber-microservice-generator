FROM golang:1.25-alpine AS builder

WORKDIR /src/generator-web

COPY generator-web/go.mod ./
RUN go mod download

COPY generator-web/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/generator-web .

FROM alpine:3.21

WORKDIR /template

RUN apk add --no-cache bash ca-certificates perl python3 rsync

COPY --from=builder /out/generator-web /usr/local/bin/generator-web
COPY auth-service ./auth-service
COPY user-service ./user-service
COPY scripts ./scripts
COPY traefik ./traefik
COPY docker-compose.yml ./docker-compose.yml

ENV REPO_ROOT=/template

EXPOSE 8090

CMD ["generator-web"]
