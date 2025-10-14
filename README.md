# win-sound-dev-go-bridge

Go + cgo bridge to `SoundAgentApiDll.dll` for monitoring/querying Windows default audio devices.

## Build (powershell)
Prereqs: CGO enabled and a GCC-style toolchain (MinGW-w64 gcc or LLVM-mingw clang). MSVC `cl.exe` is not supported by cgo.

```powershell
$Env:CGO_ENABLE=1

go build -o (Join-Path $PWD.Path 'bin/')

.\scripts\fetch-native.ps1

```


## Run
```powershell
.\bin\win-sound-dev-go-bridge.exe
```

## External module
- github.com/eduarddanziger/sound-win-scanner/v4 (pkg/soundlibwrap): cgo wrapper around SoundAgentApi, see [soundlibwrap documentation](https://pkg.go.dev/github.com/eduarddanziger/sound-win-scanner/v4/pkg/soundlibwrap)

## Advanced
Use clang instead of gcc:
```bat
set CC=x86_64-w64-mingw32-clang
set CXX=x86_64-w64-mingw32-clang++
```
