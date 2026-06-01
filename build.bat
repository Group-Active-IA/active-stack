@echo off
REM ============================================================================
REM  build.bat - JR Stack
REM  Cross-compila binarios estaticos (CGO desactivado) a la raiz del repo.
REM  Uso: doble-click, o desde terminal:  build.bat
REM  Requiere: Go 1.26+ en el PATH.
REM ============================================================================
setlocal
cd /d "%~dp0"

set "OUT=."

set "CGO_ENABLED=0"

echo == Windows amd64 ==
set "GOOS=windows" & set "GOARCH=amd64"
go build -trimpath -o "%OUT%\jr-stack_windows_amd64.exe" ./cmd/jr-stack || goto :err

echo == Linux amd64 ==
set "GOOS=linux" & set "GOARCH=amd64"
go build -trimpath -o "%OUT%\jr-stack_linux_amd64" ./cmd/jr-stack || goto :err

echo == Acceso directo en el escritorio ==
powershell -NoProfile -ExecutionPolicy Bypass -Command "$d=[Environment]::GetFolderPath('Desktop'); $exe=Join-Path '%~dp0' 'jr-stack_windows_amd64.exe'; $w=New-Object -ComObject WScript.Shell; $s=$w.CreateShortcut((Join-Path $d 'JR Stack.lnk')); $s.TargetPath=$exe; $s.WorkingDirectory=('%~dp0').TrimEnd('\'); $s.IconLocation=$exe + ',0'; $s.Description='JR Stack installer'; $s.Save(); Write-Host ('  -> ' + (Join-Path $d 'JR Stack.lnk'))" || echo   (No se pudo crear el acceso directo; el build continua igual.)

echo.
echo Listo. Binarios en la raiz del repo:
dir /b jr-stack_*
echo.
echo Tip: para otras arquitecturas (ej. ARM), cambia GOARCH=arm64 o agrega una linea.
endlocal
goto :eof

:err
echo.
echo ERROR: fallo el build. Verifica que Go este instalado y en el PATH.
endlocal
exit /b 1
