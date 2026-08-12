@echo off
title authz-engine stop
echo Stopping authz-engine server...
taskkill /f /im authz-server.exe
echo Done.
pause