#!/bin/bash
set -e

DOTNET_PATH="$(which dotnet 2>/dev/null || true)"

if [ -n "$DOTNET_PATH" ]; then
    REAL_DOTNET="$(readlink -f "$DOTNET_PATH")"
    export DOTNET_ROOT="$(dirname "$REAL_DOTNET")"
    export PATH="$DOTNET_ROOT:$HOME/.dotnet/tools:$PATH"
fi

echo "Starting..."

docker compose down -v
docker compose up -d

echo "Waiting for PostgreSQL to be ready..."
until docker exec postgres_db pg_isready > /dev/null 2>&1; do
  sleep 1
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(cd "${SCRIPT_DIR}/../backend" && pwd)"

DATETIME=$(date +"%Y%m%d%H%M")
MIGRATION_NAME="Migration_${DATETIME}"

cd "${BASE_DIR}/NetSolution/WalletService"
dotnet exec "$HOME/.dotnet/tools/.store/dotnet-ef/8.0.14/dotnet-ef/8.0.14/tools/net8.0/any/dotnet-ef.dll" migrations add "${MIGRATION_NAME}" || dotnet ef migrations add "${MIGRATION_NAME}"
dotnet exec "$HOME/.dotnet/tools/.store/dotnet-ef/8.0.14/dotnet-ef/8.0.14/tools/net8.0/any/dotnet-ef.dll" database update || dotnet ef database update

cd "${BASE_DIR}/NetSolution/TransactionService"
dotnet exec "$HOME/.dotnet/tools/.store/dotnet-ef/8.0.14/dotnet-ef/8.0.14/tools/net8.0/any/dotnet-ef.dll" migrations add "${MIGRATION_NAME}" || dotnet ef migrations add "${MIGRATION_NAME}"
dotnet exec "$HOME/.dotnet/tools/.store/dotnet-ef/8.0.14/dotnet-ef/8.0.14/tools/net8.0/any/dotnet-ef.dll" database update || dotnet ef database update

echo "Press anything to exit"
read -n 1 -s -r