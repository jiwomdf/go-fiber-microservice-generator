# Go Fiber Microservice Generator

A starter kit for building Go microservices with `Fiber`, `Traefik`

Live Generator app:
**https://jiwomdf.github.io/go-fiber-microservice-generator/**

## What It Includes

- `auth-service`
- generated extra services
- `Traefik` gateway config
- root `docker-compose.yml`
- reusable service template for future scaffolding

## Why It's Useful

- instead of starting from scratch, you may start from this starter kit
- Good for prototyping and planning
- Simple `Go Fiber` microservices with `Traefik` gateway setup

## How to Use

1. Firt open the [Live Generator App](https://jiwomdf.github.io/go-fiber-microservice-generator/), fill required services data and click `Generate`.

2. Then open the downloaded zip file and open the `docker-compose.yml` file and run the following command

```bash
docker compose up --build
```
