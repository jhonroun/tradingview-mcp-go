@echo off
REM Build tradingview-mcp-go binaries for Windows (amd64)
REM Run from repository root: scripts\build.bat

if not exist bin mkdir bin

echo Building tvmcp.exe...
go build -ldflags="-s -w" -o bin\tvmcp.exe .\cmd\tvmcp
if errorlevel 1 (echo FAILED: tvmcp & exit /b 1)

echo Building tv.exe...
go build -ldflags="-s -w" -o bin\tv.exe .\cmd\tv
if errorlevel 1 (echo FAILED: tv & exit /b 1)

echo.
echo Done: bin\tvmcp.exe  bin\tv.exe
