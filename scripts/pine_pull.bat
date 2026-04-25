@echo off
REM Pull current Pine Script source from TradingView editor → scripts\current.pine
REM Usage: scripts\pine_pull.bat

set "TV=%TV%"
if "%TV%"=="" set "TV=tv"

"%TV%" pine get > "%~dp0current.pine"
echo Pulled to scripts\current.pine
