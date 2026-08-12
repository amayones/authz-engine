@echo off
title authz-engine server
cd /d %~dp0

REM ============================================
REM  authz-engine - Self-Hosted Server
REM  Database: SQL Server (lokal)
REM  Edit konfigurasi di bawah sesuai kebutuhan
REM ============================================

REM Database: SQL Server lokal
set AUTHZ_DB_DRIVER=sqlserver
set AUTHZ_DB_CONN=sqlserver://sa:PASSWORD@localhost:1433?database=authzdb&encrypt=true&trustservercertificate=true

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