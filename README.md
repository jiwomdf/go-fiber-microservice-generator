# Go Fiber Microservice Generator

Build Go Microservices Effortlessly with a `Fiber` & `Traefik` Code Generator

Live Generator app:
**https://jiwomdf.github.io/go-fiber-microservice-generator/**

## What It Includes

- Generated microservices, including an `auth-service`
- `Traefik` gateway configuration
- A root `docker-compose.yml` for easy setup

## Why It's Useful

- Great for prototyping and planning
- Automatically generates:
  - Go Fiber service code
  - SQL migrations
  - Protobuf definitions
  - Traefik configuration
  - Dockerfiles
  - OpenAPI specifications
  - Integrated docker-compose setup
- Provides a simple microservices architecture with `Go Fiber` and a `Traefik gateway`

## How to Use

1. First open the [Live Generator App](https://jiwomdf.github.io/go-fiber-microservice-generator/), fill in the required service details, and click `Generate`.

2. Extract the downloaded ZIP file, then navigate to the project directory and run the following command:

```bash
docker compose up --build
```

## Graph of the microservices with Traefik

```mermaid
flowchart LR
    C[Client]
    T[Traefik :8000]
    A[auth-service :7704]
    G1[generated-service-1 :<port>]
    G2[generated-service-2 :<port>]

    C -->|POST /api/v1/login| T
    T -->|rewrite to /api/auth-service/v1/login| A
    A -->|JWT token| T
    T --> C

    C -->|Public auth routes\n/api/v1/auth...| T
    T -->|rewrite to /api/auth-service/v1/auth...| A
    A --> T
    T --> C

    C -->|Protected routes\n/api/v1/<entity-1>...| T
    T -->|ForwardAuth verify\nAuthorization header| A
    A -->|200 OK / 401 / 403| T
    T -->|if valid:\nrewrite to /api/generated-service-1/v1/<entity-1>...| G1
    G1 --> T
    T --> C

    C -->|Protected routes\n/api/v1/<entity-2>...| T
    T -->|ForwardAuth verify\nAuthorization header| A
    A -->|200 OK / 401 / 403| T
    T -->|if valid:\nrewrite to /api/generated-service-2/v1/<entity-2>...| G2
    G2 --> T
    T --> C

```
