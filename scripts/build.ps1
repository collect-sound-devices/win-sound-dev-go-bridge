# `scripts/build.ps1`

# go to the repo root
# go to the repo root (parent of the script directory)
Set-Location -LiteralPath $PSScriptRoot
$repoRoot = [System.IO.Directory]::GetParent($PSScriptRoot).FullName
Set-Location -LiteralPath $repoRoot

$Env:CGO_ENABLED = "1"
$Env:CC = "E:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang.exe"
$Env:CXX = "E:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang++.exe"

go build -o (Join-Path $PWD.Path 'bin/')

.\scripts\internal\fetch-native.ps1

## once more
go build -o (Join-Path $PWD.Path 'bin/')