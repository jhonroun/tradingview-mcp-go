#Requires -Version 5.1
<#
.SYNOPSIS
  Configure the tvmcp MCP server in an AI client's config file.

.PARAMETER Client
  AI client name: claude | cursor | cline | windsurf | continue | codex | gemini

.PARAMETER BinDir
  Directory containing tvmcp.exe (default: $env:LOCALAPPDATA\tvmcp)

.PARAMETER List
  List supported clients and exit.

.EXAMPLE
  .\configure-mcp.ps1 -Client claude
  .\configure-mcp.ps1 -Client cursor -BinDir "C:\tools\tvmcp"
  .\configure-mcp.ps1 -List
#>
param(
    [string]$Client  = "",
    [string]$BinDir  = "$env:LOCALAPPDATA\tvmcp",
    [switch]$List
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$TvmcpPath = Join-Path $BinDir "tvmcp.exe"

if ($List) {
    Write-Host "Supported clients:"
    Write-Host "  claude    — %APPDATA%\Claude\claude_desktop_config.json"
    Write-Host "  cursor    — %APPDATA%\Cursor\User\mcp.json"
    Write-Host "  cline     — %APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json"
    Write-Host "  windsurf  — %APPDATA%\Windsurf\User\mcp.json"
    Write-Host "  continue  — %USERPROFILE%\.continue\config.json"
    Write-Host "  codex     — %APPDATA%\OpenAI\codex.json"
    Write-Host "  gemini    — %APPDATA%\Google\gemini\settings.json"
    exit 0
}

if ($Client -eq "") {
    Write-Error "Usage: .\configure-mcp.ps1 -Client <name> [-BinDir PATH]`nRun with -List to see supported clients."
    exit 1
}

# ── helpers ───────────────────────────────────────────────────────────────────

function Ensure-Dir([string]$path) {
    $dir = Split-Path $path -Parent
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
}

function Merge-McpEntry([string]$ConfigPath, [string]$ServerName, [string]$Command) {
    Ensure-Dir $ConfigPath

    if (-not (Test-Path $ConfigPath)) {
        '{}' | Set-Content -Path $ConfigPath -Encoding UTF8
    }

    $json = Get-Content $ConfigPath -Raw | ConvertFrom-Json
    if (-not $json.PSObject.Properties['mcpServers']) {
        $json | Add-Member -NotePropertyName 'mcpServers' -NotePropertyValue ([PSCustomObject]@{})
    }
    $entry = [PSCustomObject]@{ command = $Command }
    $json.mcpServers | Add-Member -NotePropertyName $ServerName -NotePropertyValue $entry -Force
    $json | ConvertTo-Json -Depth 10 | Set-Content -Path $ConfigPath -Encoding UTF8
    Write-Host "OK  $ConfigPath -> mcpServers.$ServerName"
}

# ── client handlers ────────────────────────────────────────────────────────────

function Configure-Claude {
    $cfg = "$env:APPDATA\Claude\claude_desktop_config.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

function Configure-Cursor {
    $cfg = "$env:APPDATA\Cursor\User\mcp.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

function Configure-Cline {
    $cfg = "$env:APPDATA\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
    Write-Host "   Also add to VS Code settings.json:"
    Write-Host "   `"cline.mcpServers`": {`"tradingview`": {`"command`": `"$TvmcpPath`"}}"
}

function Configure-Windsurf {
    $cfg = "$env:APPDATA\Windsurf\User\mcp.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

function Configure-Continue {
    $cfg = "$env:USERPROFILE\.continue\config.json"
    Ensure-Dir $cfg
    if (-not (Test-Path $cfg)) {
        '{"models":[],"slashCommands":[],"mcpServers":{}}' | Set-Content -Path $cfg -Encoding UTF8
    }
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

function Configure-Codex {
    $cfg = "$env:APPDATA\OpenAI\codex.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

function Configure-Gemini {
    $cfg = "$env:APPDATA\Google\gemini\settings.json"
    Merge-McpEntry $cfg "tradingview" $TvmcpPath
}

# ── dispatch ──────────────────────────────────────────────────────────────────

Write-Host "Configuring MCP server: $TvmcpPath"
Write-Host "Client: $Client"
Write-Host ""

switch ($Client.ToLower()) {
    "claude"    { Configure-Claude   }
    "cursor"    { Configure-Cursor   }
    "cline"     { Configure-Cline    }
    "windsurf"  { Configure-Windsurf }
    "continue"  { Configure-Continue }
    "codex"     { Configure-Codex    }
    "gemini"    { Configure-Gemini   }
    default {
        Write-Error "Unknown client: $Client`nRun with -List to see supported clients."
        exit 1
    }
}

Write-Host ""
Write-Host "Done. Restart $Client to pick up the new MCP server."
Write-Host "Verify: tv status"
