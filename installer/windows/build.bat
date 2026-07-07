@echo off
rem Wrapper de build.ps1 para compilar el instalador nativo de Windows.
rem Uso: doble click, o desde terminal: build.bat [version]
rem Ejemplo: build.bat 0.2.0.0
setlocal

set "VER=%~1"
if "%VER%"=="" set "VER=0.1.0.0"

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build.ps1" -InstallerVersion %VER%
set "EXITCODE=%ERRORLEVEL%"

if %EXITCODE% neq 0 (
    echo.
    echo *** BUILD FALLIDO — revisa los errores de arriba ***
)

rem Pausa solo si se lanzo con doble click (para que la ventana no se cierre).
echo %CMDCMDLINE% | findstr /i /c:"%~nx0" >nul && pause

exit /b %EXITCODE%
