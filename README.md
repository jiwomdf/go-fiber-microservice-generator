# Go Fiber Microservice Generator

A starter kit for building Go microservices with `Fiber`, `Traefik`

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

1. Firt open the [Live Generator App](https://jiwomdf.github.io/go-fiber-microservice-generator/), fill in the required service details, and click `Generate`.

2. Extract the downloaded ZIP file, then navigate to the project directory and run the following command:

```bash
docker compose up --build
```
