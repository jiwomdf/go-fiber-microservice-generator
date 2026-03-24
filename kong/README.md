# Kong gateway

This stack runs:

- `postgres` for both services
- `auth-service` on `:7704`
- `user-service` on `:7705`
- `kong` on `:8000` with the admin API on `:8001`

## Routes

- `POST /api/v1/login` -> `auth-service`
- `POST /api/v1/auth` -> `auth-service`
- `GET|POST|PATCH|DELETE /api/v1/user...` -> `user-service`

Kong applies:

- CORS
- local rate limiting
- proxy/access logging to stdout
- JWT validation on `/api/v1/user...`

## JWT validation

Kong's JWT plugin is configured to read the `iss` claim.

`auth-service` now issues JWTs with:

- `iss=auth-service`
- `sub=<email>`
- `exp`

The shared HMAC secret is set to `dev-jwt-secret-change-me` in:

- `auth-service/config.dev.yaml`
- `kong/kong.yml`

Container-specific overrides like Postgres host and exposed service ports are injected from `docker-compose.yml`, so there is no separate duplicated `config.kong.yaml` anymore.

Change both before using this outside local development.

## Run

```bash
docker compose up --build
```

## Example flow

Create an auth record:

```bash
curl -i http://localhost:8000/api/v1/auth \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}'
```

Login:

```bash
curl -i http://localhost:8000/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}'
```

Call the protected user route:

```bash
curl -i http://localhost:8000/api/v1/user \
  -H "Authorization: Bearer <jwt>"
```
