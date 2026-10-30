@echo off
REM =====================================================================
REM  build.bat - double-click to build all the .exe files this app needs.
REM  Requires Go to be installed (https://go.dev/dl/). You only run this
REM  once, or again after the code changes. After it finishes, run
REM  start.bat to launch the demo.
REM =====================================================================
setlocal
cd /d "%~dp0"

where go >nul 2>&1
if errorlevel 1 (
  echo ERROR: Go is not installed, or not on PATH.
  echo Install it from https://go.dev/dl/  then run this file again.
  echo ^(Or ask Visvesh for the ready-made .exe files instead.^)
  pause
  exit /b 1
)

echo Building the display server...
go build -o workbench-line-display.exe .
if errorlevel 1 goto failed

echo Building the serial scanner reader...
go build -o serial-capture.exe ./serial-capture
if errorlevel 1 goto failed

echo Building the keyboard-mode scanner reader ^(backup^)...
go build -o scanner-capture.exe ./scanner-capture
if errorlevel 1 goto failed

echo.
echo ============================================================
echo  BUILD OK. Created:
echo    workbench-line-display.exe   ^(the screens^)
echo    serial-capture.exe           ^(scanner in COM-port mode^)
echo    scanner-capture.exe          ^(scanner in keyboard mode, backup^)
echo.
echo  Next: double-click start.bat to launch the demo.
echo ============================================================
pause
exit /b 0

:failed
echo.
echo BUILD FAILED. Read the message above. Make sure Go is installed
echo and you have internet the first time ^(it downloads dependencies^).
pause
exit /b 1
