@echo off
:: tradingview-mcp-go — bootstrap installer for Windows (batch wrapper)
::
:: Usage:
::   bootstrap.bat
::   bootstrap.bat claude
::   bootstrap.bat cursor "C:\tools\tvmcp"
::
:: Parameters (positional):
::   %1  Client name (optional): claude, cursor, cline, windsurf, continue, codex, gemini
::   %2  Install prefix (optional, default: %LOCALAPPDATA%\tvmcp)

setlocal

set "CLIENT=%~1"
set "PREFIX=%~2"

if "%PREFIX%"=="" set "PREFIX=%LOCALAPPDATA%\tvmcp"

:: Build PowerShell argument list
set "PS_ARGS=-Prefix ""%PREFIX%"""
if not "%CLIENT%"=="" set "PS_ARGS=%PS_ARGS% -Client ""%CLIENT%"""

powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0bootstrap.ps1" %PS_ARGS%
