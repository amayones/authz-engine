@echo off
title authz-engine - Public Tunnel
cd /d %~dp0

echo ============================================
echo  authz-engine - Public Access Tunnel
echo ============================================
echo.
echo Pilih metode tunnel:
echo   1. Cloudflare Tunnel (gratis, tanpa daftar)
echo   2. ngrok (gratis, perlu daftar)
echo.
set /p choice="Pilih (1/2): "

if "%choice%"=="1" (
    echo.
    echo Menjalankan Cloudflare Tunnel...
    echo URL publik akan muncul di sini (format: https://xxx.trycloudflare.com)
    echo.
    cloudflared tunnel --url http://localhost:8080
) else if "%choice%"=="2" (
    echo.
    echo Menjalankan ngrok...
    echo URL publik akan muncul di sini (format: https://xxx.ngrok-free.app)
    echo.
    ngrok http 8080
) else (
    echo Pilihan tidak valid.
    pause
)