@echo off

echo Starting...

docker-compose down -v
docker-compose up -d

cd ..
cd backend/ApiGateway/WalletService

for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set MIGRATION_NAME=Migration_%datetime:~0,12%
dotnet ef migrations add %MIGRATION_NAME%

dotnet ef database update

echo Press anything to exit
pause >nul