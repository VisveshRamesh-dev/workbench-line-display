@echo off
REM =====================================================================
REM  Workbench Display Controller - single-station launcher
REM
REM  Starts the Go server + the serial capture helper (reads the Zebra
REM  DS3678 over its COM port via the splitter), then opens ONE Chrome
REM  kiosk window on the SECONDARY monitor. Set MONITOR_X / MONITOR_Y to
REM  the top-left corner of the secondary monitor (Windows Display Settings).
REM =====================================================================

setlocal
cd /d "%~dp0"

REM ---- Secondary monitor top-left corner ----
REM Settings > System > Display shows each monitor's position. If the
REM secondary sits to the RIGHT of a 1920-wide primary, use X=1920, Y=0.
set MONITOR_X=1920
set MONITOR_Y=0

REM ---- Find Chrome ----
set CHROME=""
if exist "%ProgramFiles%\Google\Chrome\Application\chrome.exe" set CHROME="%ProgramFiles%\Google\Chrome\Application\chrome.exe"
if exist "%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe" set CHROME="%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe"
if exist "%LocalAppData%\Google\Chrome\Application\chrome.exe" set CHROME="%LocalAppData%\Google\Chrome\Application\chrome.exe"
if %CHROME%=="" (
    echo ERROR: Chrome not found. Install Google Chrome first.
    pause
    exit /b 1
)

REM ---- Start the Go server ----
echo Starting Workbench Display Controller server...
start "workbench-server" /MIN cmd /c "workbench-display.exe"

REM Wait for server to bind the port
timeout /t 2 /nobreak >nul

REM ---- Start the serial capture helper ----
REM The Zebra DS3678 runs in SERIAL (COM port) mode via the serial splitter.
REM serial-capture reads OUR virtual COM port (set "serial_port" in config.json)
REM and forwards each scan to the display. The kiosk does NOT need focus for
REM scans to land, because they arrive over the port, not the keyboard.
if exist "serial-capture.exe" (
    echo Starting serial capture helper ^(reads the DS3678 COM port^)...
    start "serial-capture" /MIN cmd /c "serial-capture.exe"
) else (
    echo WARNING: serial-capture.exe not found - build it with:
    echo    go build -o serial-capture.exe ./serial-capture
    pause
)

set BASE=http://localhost:8080

REM ---- Open the single kiosk display on the secondary monitor ----
REM --kiosk gives full-screen with no browser chrome. --window-position
REM places it on the monitor whose top-left corner matches MONITOR_X/Y.
set FLAGS=--kiosk --disable-pinch --overscroll-history-navigation=0 --disable-features=TranslateUI --no-first-run --disable-session-crashed-bubble

echo Launching display on secondary monitor at %MONITOR_X%,%MONITOR_Y% ...
start "" %CHROME% %FLAGS% --user-data-dir="%~dp0chrome-profile-1" --window-position=%MONITOR_X%,%MONITOR_Y% %BASE%/station/WB-01

echo.
echo Display launched. The server console is minimized.
echo Scans arrive over the serial port, so the kiosk need not stay focused.
echo   Supervisor view: %BASE%/supervisor
echo   Simulator:       %BASE%/simulator

endlocal
