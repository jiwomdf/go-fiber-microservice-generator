#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/scaffold-service.sh <entity-name> [output-dir]

Example:
  ./scripts/scaffold-service.sh asset ./asset-service
EOF
}

[[ "${1:-}" =~ ^(-h|--help)?$ ]] && usage && exit 0

command -v rsync >/dev/null 2>&1 || { echo "rsync is required" >&2; exit 1; }
command -v perl >/dev/null 2>&1 || { echo "perl is required" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 1; }

slugify() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | perl -pe 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g'
}

pascal_case() {
  printf '%s' "$1" | perl -pe 's/[^A-Za-z0-9]+/ /g; s/(^| )([A-Za-z0-9])/$1 . uc($2)/ge; s/ //g'
}

replace_all() {
  local file="$1"
  local from="$2"
  local to="$3"
  FROM="$from" TO="$to" perl -0pi -e 's/\Q$ENV{FROM}\E/$ENV{TO}/g' "$file"
}

next_service_ports() {
  python3 - "$repo_root" <<'PY'
import pathlib
import re
import sys

repo_root = pathlib.Path(sys.argv[1])
http_ports = []
grpc_ports = []

for cfg in repo_root.glob("*-service/config.dev.yaml"):
    text = cfg.read_text()
    for match in re.finditer(r'^\s*port:\s*"(\d+)"\s*$', text, re.MULTILINE):
        port = int(match.group(1))
        if 7000 <= port < 8000:
            http_ports.append(port)
    grpc_match = re.search(r'^\s*grpc_port:\s*"(\d+)"\s*$', text, re.MULTILINE)
    if grpc_match:
        grpc_ports.append(int(grpc_match.group(1)))

next_http = max(http_ports, default=7703) + 1
next_grpc = max(grpc_ports, default=57703) + 1
print(f"{next_http} {next_grpc}")
PY
}

append_compose_service() {
  local compose_file="$repo_root/docker-compose.yml"
  [[ -f "$compose_file" ]] || return 0
  grep -q "^  ${service_name}:" "$compose_file" && return 0

  SERVICE_NAME="$service_name" ENTITY="$entity" HTTP_PORT="$http_port" python3 - "$compose_file" <<'PY'
import os
import pathlib
import sys

compose_file = pathlib.Path(sys.argv[1])
service_name = os.environ["SERVICE_NAME"]
entity = os.environ["ENTITY"]
http_port = os.environ["HTTP_PORT"]

block = f"""
  {service_name}:
    build:
      context: ./{service_name}
    container_name: {service_name}
    environment:
      POSTGRES_HOST: host.docker.internal
      POSTGRES_PORT: "5432"
      POSTGRES_DATABASE: belajarmudah
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ""
      POSTGRES_SCHEMA: public
      SERVER_HTTP_PORT: "{http_port}"
    ports:
      - "{http_port}:{http_port}"
    labels:
      - "traefik.enable=true"
      - "traefik.http.services.{service_name}.loadbalancer.server.port={http_port}"
      - "traefik.http.routers.{entity}-api.rule=PathPrefix(`/api/v1/{entity}`)"
      - "traefik.http.routers.{entity}-api.entrypoints=web"
      - "traefik.http.routers.{entity}-api.middlewares={entity}-strip-v1,{entity}-addprefix,protected-common@file"
      - "traefik.http.routers.{entity}-api.service={service_name}"
      - "traefik.http.middlewares.{entity}-strip-v1.stripprefix.prefixes=/api/v1"
      - "traefik.http.middlewares.{entity}-addprefix.addprefix.prefix=/api/{service_name}/v1"
"""

text = compose_file.read_text()
marker = "\n  traefik:\n"
if marker not in text:
    raise SystemExit("could not find traefik service block in docker-compose.yml")
text = text.replace(marker, block + marker, 1)
compose_file.write_text(text)
PY
}

entity="$(slugify "${1:-}")"
[[ -n "$entity" ]] || { echo "entity name must contain letters or numbers" >&2; exit 1; }

entity_pascal="$(pascal_case "$entity")"
entity_plural="${entity}s"
entity_plural_pascal="${entity_pascal}s"
service_name="${entity}-service"
entity_upper="$(printf '%s' "$entity" | tr '[:lower:]' '[:upper:]')"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
source_dir="$repo_root/user-service"
output_dir="${2:-$repo_root/$service_name}"
target_dir="$(cd "$(dirname "$output_dir")" && pwd)/$(basename "$output_dir")"

[[ -d "$source_dir" ]] || { echo "template directory not found: $source_dir" >&2; exit 1; }
[[ ! -e "$target_dir" ]] || { echo "target directory already exists: $target_dir" >&2; exit 1; }

read -r http_port grpc_port < <(next_service_ports)

mkdir -p "$(dirname "$target_dir")"
rsync -a \
  --exclude '.git' \
  --exclude '.idea' \
  --exclude '.vscode' \
  --exclude '.DS_Store' \
  --exclude 'app/__debug_bin*' \
  --exclude 'scripts/scaffold-service.sh' \
  "$source_dir/" "$target_dir/"

mv "$target_dir/domain/user.go" "$target_dir/domain/${entity}.go"
mv "$target_dir/data/usecase/user.go" "$target_dir/data/usecase/${entity}.go"
mv "$target_dir/data/repository/postgres/user.go" "$target_dir/data/repository/postgres/${entity}.go"
mv "$target_dir/data/delivery/http/handler/user.go" "$target_dir/data/delivery/http/handler/${entity}.go"
mv "$target_dir/data/delivery/grpc/proto/user.proto" "$target_dir/data/delivery/grpc/proto/${entity}.proto"

files=(
  "$target_dir/go.mod"
  "$target_dir/Dockerfile"
  "$target_dir/Makefile"
  "$target_dir/app/main.go"
  "$target_dir/config.dev.yaml"
  "$target_dir/config.local.yaml"
  "$target_dir/config/struct.go"
  "$target_dir/domain/errors.go"
  "$target_dir/domain/${entity}.go"
  "$target_dir/data/usecase/${entity}.go"
  "$target_dir/data/repository/postgres/${entity}.go"
  "$target_dir/data/delivery/http/router.go"
  "$target_dir/data/delivery/http/handler/${entity}.go"
  "$target_dir/data/delivery/grpc/proto/${entity}.proto"
  "$target_dir/migration/01_initial_schema.up.sql"
  "$target_dir/openapi.yaml"
)

for file in "${files[@]}"; do
  [[ -f "$file" ]] || continue
  replace_all "$file" "user-service" "$service_name"
  replace_all "$file" "CreateUserRequest" "Create${entity_pascal}Request"
  replace_all "$file" "UpdateUserRequest" "Update${entity_pascal}Request"
  replace_all "$file" "GetUserByIdRequest" "Get${entity_pascal}ByIdRequest"
  replace_all "$file" "DeleteUserRequest" "Delete${entity_pascal}Request"
  replace_all "$file" "DeleteUserResponse" "Delete${entity_pascal}Response"
  replace_all "$file" "UserListResponse" "${entity_pascal}ListResponse"
  replace_all "$file" "UserResponse" "${entity_pascal}Response"
  replace_all "$file" "UserHandler" "${entity_pascal}Handler"
  replace_all "$file" "UserUsecase" "${entity_pascal}Usecase"
  replace_all "$file" "UserRepository" "${entity_pascal}Repository"
  replace_all "$file" "UserRepo" "${entity_pascal}Repo"
  replace_all "$file" "UserService" "${entity_pascal}Service"
  replace_all "$file" "NewUserHandler" "New${entity_pascal}Handler"
  replace_all "$file" "NewUserUsecase" "New${entity_pascal}Usecase"
  replace_all "$file" "NewUserRepo" "New${entity_pascal}Repo"
  replace_all "$file" "GetAllUsers" "GetAll${entity_plural_pascal}"
  replace_all "$file" "GetUserById" "Get${entity_pascal}ById"
  replace_all "$file" "CreateUser" "Create${entity_pascal}"
  replace_all "$file" "UpdateUser" "Update${entity_pascal}"
  replace_all "$file" "DeleteUser" "Delete${entity_pascal}"
  replace_all "$file" "/user" "/$entity"
  replace_all "$file" "\"users\"" "\"${entity_plural}\""
  replace_all "$file" " public.users " " public.${entity_plural} "
  replace_all "$file" "user endpoints" "${entity} endpoints"
  replace_all "$file" "Create user" "Create ${entity}"
  replace_all "$file" "Get all users" "Get all ${entity_plural}"
  replace_all "$file" "Get user by ID" "Get ${entity} by ID"
  replace_all "$file" "Update user by ID" "Update ${entity} by ID"
  replace_all "$file" "Delete user by ID" "Delete ${entity} by ID"
  replace_all "$file" "User Routes" "${entity_pascal} Routes"
done

entity_files=(
  "$target_dir/domain/errors.go"
  "$target_dir/domain/${entity}.go"
  "$target_dir/data/usecase/${entity}.go"
  "$target_dir/data/repository/postgres/${entity}.go"
  "$target_dir/data/delivery/http/handler/${entity}.go"
  "$target_dir/data/delivery/grpc/proto/${entity}.proto"
  "$target_dir/openapi.yaml"
)

for file in "${entity_files[@]}"; do
  [[ -f "$file" ]] || continue
  replace_all "$file" "Users" "$entity_plural_pascal"
  replace_all "$file" "User" "$entity_pascal"
done

replace_all "$target_dir/config/struct.go" 'ClientID:          "user",' "ClientID:          \"${entity}\","
replace_all "$target_dir/config.dev.yaml" 'port: "7705"' "port: \"${http_port}\""
replace_all "$target_dir/config.local.yaml" 'port: "7705"' "port: \"${http_port}\""
replace_all "$target_dir/config.dev.yaml" 'grpc_port: "57705"' "grpc_port: \"${grpc_port}\""
replace_all "$target_dir/config.local.yaml" 'grpc_port: "57705"' "grpc_port: \"${grpc_port}\""
replace_all "$target_dir/Dockerfile" 'EXPOSE 7705' "EXPOSE ${http_port}"
replace_all "$target_dir/openapi.yaml" 'url: http://localhost:7704/api/' "url: http://localhost:${http_port}/api/"
replace_all "$target_dir/domain/errors.go" 'const StatusCodePrefix = "USER"' "const StatusCodePrefix = \"${entity_upper}\""

if [[ "$target_dir" == "$repo_root/"* ]]; then
  append_compose_service
fi

echo "Generated service at: $target_dir"
echo "Assigned ports: http=${http_port}, grpc=${grpc_port}"
