@echo off
rem Lanzador de Windows del validador del modo equipo: elige un Python >= 3.8.
rem El chequeo de version descarta el alias falso de Microsoft Store.
setlocal
set "PY="
for %%C in ("py -3" python3 python) do (
  if not defined PY (
    %%~C -c "import sys; sys.exit(0 if sys.version_info >= (3, 8) else 1)" >nul 2>&1 && set "PY=%%~C"
  )
)
if not defined PY (
  echo {"ok": false, "errors": ["falta Python 3.8 o superior para validar team.json"]}
  exit /b 3
)
%PY% -B "%~dp0team.py" %*
exit /b %ERRORLEVEL%
