@echo off
REM Push scripts\current.pine into TradingView Pine editor and compile.
REM Usage: scripts\pine_push.bat

set "TV=%TV%"
if "%TV%"=="" set "TV=tv"
set "PINE_FILE=%~dp0current.pine"

if not exist "%PINE_FILE%" (
  echo Error: %PINE_FILE% not found. Run pine_pull.bat first.
  exit /b 1
)

set /p SRC=<"%PINE_FILE%"
"%TV%" pine set "%SRC%"
"%TV%" pine smart-compile
"%TV%" pine errors
