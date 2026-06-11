@echo off
REM ============================================================================
REM  build.bat - Active Stack
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
go build -trimpath -o "%OUT%\active-stack_windows_amd64.exe" ./cmd/active-stack || goto :err

echo == Linux amd64 ==
set "GOOS=linux" & set "GOARCH=amd64"
go build -trimpath -o "%OUT%\active-stack_linux_amd64" ./cmd/active-stack || goto :err

echo == Acceso directo en el escritorio ==
powershell -NoProfile -ExecutionPolicy Bypass -Command "$d=[Environment]::GetFolderPath('Desktop'); $exe=Join-Path '%~dp0' 'active-stack_windows_amd64.exe'; $w=New-Object -ComObject WScript.Shell; $s=$w.CreateShortcut((Join-Path $d 'Active Stack.lnk')); $s.TargetPath=$exe; $s.WorkingDirectory=('%~dp0').TrimEnd('\'); $s.IconLocation=$exe + ',0'; $s.Description='Active Stack installer'; $s.Save(); Write-Host ('  -> ' + (Join-Path $d 'Active Stack.lnk'))" || echo   (No se pudo crear el acceso directo; el build continua igual.)

echo.
echo Listo. Binarios en la raiz del repo:
dir /b active-stack_*
echo.
echo Tip: para otras arquitecturas (ej. ARM), cambia GOARCH=arm64 o agrega una linea.
echo.
echo Siguiente paso: copia el binario a un directorio en tu PATH y ejecuta:
echo   active-stack install
echo Esto instala los harnesses Y registra el binario en tu PATH automaticamente.
endlocal
goto :eof

:err
echo.
echo ERROR: fallo el build. Verifica que Go este instalado y en el PATH.
endlocal
exit /b 1
