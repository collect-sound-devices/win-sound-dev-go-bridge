# PowerShell
<#!
Usage (local dev):
  .\scripts\fetch-native.ps1 -Tag v4.0.4-rc002 \
    -RepoPath = "github.com/eduarddanziger/sound-win-scanner"
    -PkgSubDir = "github.com\eduarddanziger\sound-win-scanner\v4@v4.0.4-rc002\pkg\soundlibwrap" \
    -DllOutDir .\out

This version downloads the release asset directly via HTTPS:
  https://github.com/eduarddanziger/sound-win-scanner/releases/download/<Tag>/SoundAgentApi-<Tag>.zip
#>

param(
  [Parameter(Mandatory = $true)][string]$Tag,                         # e.g. v1.2.3
  [string]$RepoPath = "github.com/eduarddanziger/sound-win-scanner",
  [string]$PkgSubDir = "pkg\soundlibwrap",
  [string]$DllOutDir = "E:\DWP\github\win-sound-dev-go-bridge\bin"
)

$ErrorActionPreference = "Stop"

$baseModDir = (go env GOMODCACHE)

$PkgDir = Join-Path -Path $baseModDir -ChildPath $RepoPath
$PkgDir = Join-Path -Path $PkgDir -ChildPath "v4@$Tag"
$PkgDir = Join-Path -Path $PkgDir -ChildPath $PkgSubDir

Write-Host "PkgSubDir: $PkgSubDir ..."
Write-Host "DllOutDir: $DllOutDir ..."


if ($DllOutDir) {
    New-Item -ItemType Directory -Force -Path $DllOutDir | Out-Null
}

# Ensure destination dirs
New-Item -ItemType Directory -Force -Path $PkgDir | Out-Null


# Temp dirs
$tmp = Join-Path $env:TEMP ("ghrel_" + [guid]::NewGuid())
$null = New-Item -ItemType Directory -Force -Path $tmp
$extractDir = Join-Path $tmp "extracted"
$null = New-Item -ItemType Directory -Force -Path $extractDir

# Download the zipped asset directly via HTTPS
$zipName = "SoundAgentApi-$Tag.zip"
$zipUrl  = "https://" + $RepoPath + "/releases/download/$Tag/$zipName"
$zipPath = Join-Path $tmp $zipName

Write-Host "Downloading $zipUrl ..."

try {
    Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath -UseBasicParsing -ErrorAction Stop
}
catch {
    throw "Failed to download $zipUrl : $($_.Exception.Message)"
}

if (-not (Test-Path $zipPath)) { throw "Downloaded file not found: $zipPath" }
if ((Get-Item $zipPath).Length -lt 100) { Write-Warning "Zip file size is unexpectedly small (<100 bytes)." }

# Extract
Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

# Locate required files inside the extracted tree
$h   = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApi.h" | Select-Object -First 1
$lib = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApi.lib" | Select-Object -First 1
$dll = Get-ChildItem -Path $extractDir -Recurse -Filter "SoundAgentApi.dll" | Select-Object -First 1

if (-not $h)   { throw "Header 'SoundAgentApi.h' not found in the zip." }
if (-not $lib) { throw "Library 'SoundAgentApi.lib' not found in the zip." }
if (-not $dll) { throw "DLL 'SoundAgentApi.dll' not found in the zip." }

# Copy to module layout
Copy-Item $h.FullName   -Destination (Join-Path $PkgDir "SoundAgentApi.h") -Force
Copy-Item $lib.FullName -Destination (Join-Path $PkgDir "SoundAgentApi.lib") -Force

if ($DllOutDir) {
    Copy-Item $dll.FullName -Destination (Join-Path $DllOutDir "SoundAgentApi.dll") -Force
}

Write-Host "Native assets synced from '$zipName':"
Write-Host "  .h   -> '$PkgDir'"
Write-Host "  .lib -> '$PkgDir'"
if ($DllOutDir) {
    Write-Host "  .dll -> '$DllOutDir'"
} else {
    Write-Host "  .dll (not copied - DllOutDir not set)"
}

# Cleanup
Remove-Item -Recurse -Force $tmp