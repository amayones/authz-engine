@echo off
title authz-engine server
cd /d %~dp0

REM ============================================
REM  authz-engine - Self-Hosted Server
REM  Database: SQL Server (lokal)
REM  User: may  Password: may
REM ============================================

REM Build dulu jika belum ada
if not exist authz-server.exe (
    echo Building authz-server.exe...
    go build -o authz-server.exe ./cmd/server
    if errorlevel 1 (
        echo Build gagal!
        pause
        exit /b 1
    )
)

REM Database: SQL Server lokal
REM Catatan: ^& digunakan untuk escape karakter & di batch file
set AUTHZ_DB_DRIVER=sqlserver
set AUTHZ_DB_CONN=sqlserver://may:may@localhost:1433?database=authzdb^&encrypt=true^&trustservercertificate=true

REM HTTP server
set AUTHZ_ADDR=:8080

REM Auto migration saat startup
set AUTHZ_AUTO_MIGRATE=true

echo.
echo ============================================
echo  authz-engine server
echo  DB: %AUTHZ_DB_DRIVER%
echo  Addr: %AUTHZ_ADDR%
echo  AutoMigrate: %AUTHZ_AUTO_MIGRATE%
echo ============================================
echo.

authz-server.exe
pause