@echo off
title authz-engine server
cd /d %~dp0

REM ============================================
REM  authz-engine - Self-Hosted Server
REM  Edit konfigurasi di bawah sesuai kebutuhan
REM ============================================

REM Database: PostgreSQL lokal atau Supabase
set AUTHZ_DB_DRIVER=postgres
set AUTHZ_DB_CONN=postgresql://postgres:PASSWORD@localhost:5432/authzdb

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