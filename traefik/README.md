# Traefik gateway

This stack runs:

- `auth-service` on `:7704`
- `user-service` on `:7705`
- `traefik` on `:8000`

For local Docker on macOS, both services connect to your existing host Postgres via `host.docker.internal`.

## Routes

- `POST /api/v1/login` -> `auth-service`
- `GET|POST|PATCH|DELETE /api/v1/auth...` -> `auth-service`
- `GET|POST|PATCH|DELETE /api/v1/user...` -> `user-service`

Traefik applies:

- CORS
- rate limiting
- access logging
- `ForwardAuth` protection on `/api/v1/user...`

## Auth flow

`auth-service` issues JWTs and exposes `GET /api/auth-service/v1/verify`.

Traefik calls that verify endpoint before forwarding protected `user-service` requests. On success, Traefik forwards these headers upstream:

- `X-Auth-User`
- `X-Auth-Email`
- `X-Auth-Subject`
- `X-Auth-Issuer`
- `X-Auth-Token-Type`

## Run

```bash
docker compose up --build
```

Expected local DB settings from Docker Compose:

- host: `host.docker.internal`
- port: `5432`
- database: `belajarmudah`
- user: `postgres`
- password: empty
