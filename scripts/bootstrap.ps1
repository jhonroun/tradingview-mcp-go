#Requires -Version 5.1
<#
.SYNOPSIS
  tradingview-mcp-go — bootstrap installer for Windows.

.DESCRIPTION
  Downloads and installs tvmcp.exe + tv.exe from GitHub Releases.
  Optionally configures the MCP server in an AI client.

.EXAMPLE
  # One-line install (PowerShell):
  iwr -useb https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.ps1 | iex

  # With options (save and run):
  .\bootstrap.ps1 -Client claude
  .\bootstrap.ps1 -Client cursor -Prefix "C:\tools\tvmcp"
  .\bootstrap.ps1 -Version "v1.2.0"
  .\bootstrap.ps1 -ListClients

.PARAMETER Client
  AI client to configure after install: claude, cursor, cline, windsurf, continue, codex, gemini

.PARAMETER Prefix
  Install directory for binaries (default: $env:LOCALAPPDATA\tvmcp)

.PARAMETER Version
  Release tag to install (default: latest)

.PARAMETER ListClients
  List supported client names and exit.

.PARAMETER NoPath
  Skip adding Prefix to the user PATH.

.PARAMETER NoConfigure
  Skip MCP client configuration even if -Client is specified.
#>
param(
    [string]$Client       = "",
    [string]$Prefix       = "$env:LOCALAPPDATA\tvmcp",
    [string]$Version      = "latest",
    [switch]$ListClients,
    [switch]$NoPath,
    [switch]$NoConfigure
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Repo    = "jhonroun/tradingview-mcp-go"
$GhApi   = "https://api.github.com/repos/$Repo"
$GhRaw   = "https://raw.githubusercontent.com/$Repo/main"

if ($ListClients) {
    Write-Host "claude cursor cline windsurf continue codex gemini"
    exit 0
}

# ── detect arch ───────────────────────────────────────────────────────────────

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { "amd64" }   # safe fallback
}

$Platform = "windows-$Arch"

Write-Host "=== tradingview-mcp-go installer (Windows) ==="
Write-Host "Platform : $Platform"
Write-Host "Prefix   : $Prefix"

# ── resolve version ───────────────────────────────────────────────────────────

if ($Version -eq "latest") {
    Write-Host -NoNewline "Fetching latest release... "
    $rel     = Invoke-RestMethod "$GhApi/releases/latest"
    $Version = $rel.tag_name
    Write-Host $Version
}

# ── download ──────────────────────────────────────────────────────────────────

$Archive    = "tradingview-mcp-go_${Version}_${Platform}"
$ArchiveFile = "$Archive.zip"
$Url         = "https://github.com/$Repo/releases/download/$Version/$ArchiveFile"

$TmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null
try {

Write-Host "Downloading $ArchiveFile ..."
$ZipPath = Join-Path $TmpDir $ArchiveFile
Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing

# ── extract ───────────────────────────────────────────────────────────────────

Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force
$PkgDir = Join-Path $TmpDir $Archive

# ── install binaries ──────────────────────────────────────────────────────────

if (-not (Test-Path $Prefix)) {
    New-Item -ItemType Directory -Path $Prefix -Force | Out-Null
}

Copy-Item "$PkgDir\tvmcp.exe" "$Prefix\tvmcp.exe" -Force
Copy-Item "$PkgDir\tv.exe"    "$Prefix\tv.exe"    -Force
Write-Host "Installed: $Prefix\tvmcp.exe  $Prefix\tv.exe"

# ── add to PATH ───────────────────────────────────────────────────────────────

if (-not $NoPath) {
    $CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($CurrentPath -notlike "*$Prefix*") {
        [Environment]::SetEnvironmentVariable("PATH", "$CurrentPath;$Prefix", "User")
        Write-Host "Added $Prefix to user PATH (restart shell to take effect)"
    }
}

# ── install agents + skills + prompts ─────────────────────────────────────────

$ShareDir = Join-Path (Split-Path $Prefix -Parent) "tradingview-mcp-go"
if (Test-Path "$PkgDir\agents") {
    New-Item -ItemType Directory -Path $ShareDir -Force | Out-Null
    Copy-Item "$PkgDir\agents"  $ShareDir -Recurse -Force
    Copy-Item "$PkgDir\skills"  $ShareDir -Recurse -Force
    Copy-Item "$PkgDir\prompts" $ShareDir -Recurse -Force
    Write-Host "Assets:    $ShareDir\{agents,skills,prompts}"
}

# ── configure MCP client ──────────────────────────────────────────────────────

if ($Client -ne "" -and -not $NoConfigure) {
    Write-Host ""
    Write-Host "Configuring MCP for client: $Client"
    $ConfScript = "$PkgDir\scripts\configure-mcp.ps1"
    if (-not (Test-Path $ConfScript)) {
        # Download fallback
        $ConfScript = Join-Path $TmpDir "configure-mcp.ps1"
        Invoke-WebRequest "$GhRaw/scripts/configure-mcp.ps1" -OutFile $ConfScript -UseBasicParsing
    }
    & $ConfScript -Client $Client -BinDir $Prefix
}

} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}

# ── done ─────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "=== Installation complete ==="
Write-Host "  tv status      — verify CDP connection"
Write-Host "  tv launch      — start TradingView with CDP"
Write-Host ""
if ($Client -ne "") {
    Write-Host "MCP configured for: $Client"
    Write-Host "Restart $Client to activate the tradingview MCP server."
} else {
    Write-Host "To configure MCP for your client:"
    Write-Host "  .\configure-mcp.ps1 -Client <claude|cursor|cline|windsurf|continue|codex|gemini>"
}
