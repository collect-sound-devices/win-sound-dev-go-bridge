<#
.DESCRIPTION
Build Windows sound binaries and fetch native deps; use -mingwPath (alias -m) to point at an llvm-mingw root, or set it to an empty string to leave CC/CXX unchanged.
.PARAMETER mingwPath
Path to llvm-mingw root; set to empty to skip overriding CC/CXX.
#>

Param(
    [Alias("m")]
    [Parameter(HelpMessage = "Path to llvm-mingw root; set to empty to skip overriding CC/CXX.")]
    [string]$mingwPath = "E:\\tools\\llvm-mingw\\"
)

# go to the repo root
# go to the repo root (parent of the script directory)
Set-Location -LiteralPath $PSScriptRoot
$repoRoot = [System.IO.Directory]::GetParent($PSScriptRoot).FullName
Set-Location -LiteralPath $repoRoot

$Env:CGO_ENABLED = "1"
if ($mingwPath -ne "") {
    if (-not (Test-Path -LiteralPath $mingwPath)) {
        Write-Error "mingwPath '$mingwPath' does not exist. Set it to a valid llvm-mingw root or pass an empty string to skip overriding CC/CXX."
        Get-Help $PSCommandPath -Detailed
        exit 1
    }
    $Env:CC = Join-Path $mingwPath "bin/x86_64-w64-mingw32-clang.exe"
    $Env:CXX = Join-Path $mingwPath "bin/x86_64-w64-mingw32-clang++.exe"
}

go build -v -o (Join-Path $PWD.Path 'bin/win-sound-scanner.exe') ./cmd/win-sound-scanner

.\scripts\internal\fetch-native.ps1

## once more
go build -v -o (Join-Path $PWD.Path 'bin/win-sound-scanner.exe') ./cmd/win-sound-scanner
