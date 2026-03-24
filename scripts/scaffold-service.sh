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

entity="$(slugify "${1:-}")"
[[ -n "$entity" ]] || { echo "entity name must contain letters or numbers" >&2; exit 1; }

entity_pascal="$(pascal_case "$entity")"
entity_plural="${entity}s"
entity_plural_pascal="${entity_pascal}s"
service_name="${entity}-service"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
source_dir="$repo_root/user-service"
output_dir="${2:-$repo_root/$service_name}"
target_dir="$(cd "$(dirname "$output_dir")" && pwd)/$(basename "$output_dir")"

[[ -d "$source_dir" ]] || { echo "template directory not found: $source_dir" >&2; exit 1; }
[[ ! -e "$target_dir" ]] || { echo "target directory already exists: $target_dir" >&2; exit 1; }

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

echo "Generated service at: $target_dir"
