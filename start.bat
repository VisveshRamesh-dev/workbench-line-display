@echo off
REM =====================================================================
REM  Workbench Line Display - multi-monitor pipeline launcher
REM
REM  Starts the pipeline server + the serial reader, then opens ONE Chrome
REM  kiosk per monitor. Set each monitor's top-left X,Y from Windows
REM  Display Settings. Add or remove a "Monitor" block below to match your
REM  number of screens - it must match the "displays" list in config.json.
REM  The PRIMARY monitor (plant's software) is NOT listed here.
REM =====================================================================
setlocal
cd /d "%~dp0"

set CHROME=""
if exist "%ProgramFiles%\Google\Chrome\Application\chrome.exe" set CHROME="%ProgramFiles%\Google\Chrome\Application\chrome.exe"
if exist "%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe" set CHROME="%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe"
if exist "%LocalAppData%\Google\Chrome\Application\chrome.exe" set CHROME="%LocalAppData%\Google\Chrome\Application\chrome.exe"
if %CHROME%=="" ( echo ERROR: Chrome not found. Install Google Chrome first. & pause & exit /b 1 )

echo Starting pipeline server...
start "line-server" /MIN cmd /c "workbench-line-display.exe"
timeout /t 2 /nobreak >nul

if exist "serial-capture.exe" (
  echo Starting serial reader...
  start "serial-capture" /MIN cmd /c "serial-capture.exe"
) else (
  echo NOTE: serial-capture.exe not found - build it with:
  echo    go build -o serial-capture.exe ./serial-capture
)

set BASE=http://localhost:8080
set FLAGS=--kiosk --disable-pinch --overscroll-history-navigation=0 --disable-features=TranslateUI --no-first-run --disable-session-crashed-bubble

REM ===== Monitor 1 (bench 1, entry) =====  EDIT the X,Y to this monitor's top-left
start "" %CHROME% %FLAGS% --user-data-dir="%~dp0chrome-profile-D1" --window-position=1920,0 %BASE%/display/D1

REM ===== Monitor 2 (bench 2) =====
start "" %CHROME% %FLAGS% --user-data-dir="%~dp0chrome-profile-D2" --window-position=3000,0 %BASE%/display/D2

REM ===== Monitor 3 (bench 3) =====
start "" %CHROME% %FLAGS% --user-data-dir="%~dp0chrome-profile-D3" --window-position=4080,0 %BASE%/display/D3

echo.
echo Pipeline launched - one kiosk per monitor.
echo   Supervisor: %BASE%/supervisor     Simulator: %BASE%/simulator
echo Set each --window-position above to the matching monitor's top-left
echo (run the screen-list PowerShell from the guide to get each X,Y).
endlocal
