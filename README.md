# win-sound-dev-go-bridge

Go + cgo bridge to `SoundAgentApiDll.dll` for monitoring/querying Windows default audio devices.

## Build (powershell)

Prereqs: CGO enabled and a GCC-style toolchain (MinGW-w64 gcc or LLVM-mingw clang).
- Download an x86_64 LLVM‑mingw build (zip) from the official releases (search for “llvm-mingw releases”).
- Your download's name is similar to llvm-mingw-20251118-msvcrt-x86_64.zip
- Copy its bin, include,lib and x86_64-w64-mingw32 folders to some folder, e.g. E:\tools\llvm-mingw 

```powershell
$Env:CGO_ENABLED = "1"
$Env:CC = "E:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang.exe"
$Env:CXX = "E:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang++.exe"

go build -o (Join-Path $PWD.Path 'bin/')

.\scripts\fetch-native.ps1

## once more
go build -o (Join-Path $PWD.Path 'bin/') 
```

## Run
```powershell
.\bin\win-sound-dev-go-bridge.exe
```

## External module
- github.com/eduarddanziger/sound-win-scanner/v4 (pkg/soundlibwrap): cgo wrapper around SoundAgentApi, see [soundlibwrap documentation](https://pkg.go.dev/github.com/eduarddanziger/sound-win-scanner/v4/pkg/soundlibwrap)

## Advanced
You can use  clang instead of gcc:
```bat
set CC=...x86_64-w64-mingw32-clang
set CXX=...x86_64-w64-mingw32-clang++
```
