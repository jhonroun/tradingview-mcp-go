@echo off
REM Install tvmcp.exe and tv.exe to %SystemRoot%\System32 (or a custom path).
REM Run as Administrator: scripts\install.bat
REM Custom prefix:        scripts\install.bat "C:\Users\You\bin"

set "DEST=%~1"
if "%DEST%"=="" set "DEST=%SystemRoot%\System32"

echo Building...
call "%~dp0build.bat"
if errorlevel 1 exit /b 1

echo Installing to %DEST%...
copy /Y bin\tvmcp.exe "%DEST%\tvmcp.exe" >nul
copy /Y bin\tv.exe    "%DEST%\tv.exe"    >nul

echo Done. Verify: tv status
