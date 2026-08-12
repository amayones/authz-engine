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

if "%choice%"=="1" goto cloudflare
if "%choice%"=="2" goto ngrok
echo Pilihan tidak valid.
pause
exit /b

:cloudflare
echo.
echo Cek cloudflared...
where cloudflared >nul 2>nul
if %errorlevel%==0 (
    echo Menjalankan Cloudflare Tunnel...
    echo URL publik akan muncul di sini (format: https://xxx.trycloudflare.com)
    echo.
    cloudflared tunnel --url http://localhost:8080
) else (
    echo.
    echo [ERROR] cloudflared tidak ditemukan!
    echo.
    echo Cara install:
    echo   1. Download dari: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
    echo   2. Pilih file: cloudflared-windows-amd64.exe
    echo   3. Rename menjadi: cloudflared.exe
    echo   4. Letakkan di folder proyek ini: %~dp0
    echo   5. Jalankan tunnel.bat lagi
    echo.
)
pause
exit /b

:ngrok
echo.
echo Cek ngrok...
where ngrok >nul 2>nul
if %errorlevel%==0 (
    echo Menjalankan ngrok...
    echo URL publik akan muncul di sini (format: https://xxx.ngrok-free.app)
    echo.
    ngrok http 8080
) else (
    echo.
    echo [ERROR] ngrok tidak ditemukan!
    echo.
    echo Cara install:
    echo   1. Download dari: https://ngrok.com/download
    echo   2. Daftar di https://ngrok.com (gratis, email)
    echo   3. Dapatkan token di dashboard
    echo   4. Jalankan: ngrok config add-authtoken YOUR_TOKEN
    echo   5. Letakkan ngrok.exe di folder proyek ini: %~dp0
    echo   6. Jalankan tunnel.bat lagi
    echo.
)
pause
exit /b