#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="${SCRIPT_DIR}/../backend/ApiGateway"

echo "Applying EF Core migrations..."
cd "${SCRIPT_DIR}"
docker compose down -v
docker compose up -d

DATETIME=$(date +"%Y%m%d%H%M")
MIGRATION_NAME="Migration_${DATETIME}"

BASE_DIR="$(pwd)/../backend/ApiGateway"

cd "${BASE_DIR}/WalletService"

dotnet ef migrations add "${MIGRATION_NAME}"
dotnet ef database update

cd "${BASE_DIR}/TransactionService"

dotnet ef migrations add "${MIGRATION_NAME}"
dotnet ef database update

echo "Press anything to exit"
read -r