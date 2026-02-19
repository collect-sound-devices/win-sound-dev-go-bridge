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

go build -o (Join-Path $PWD.Path 'bin/win-sound-scanner.exe') ./cmd/win-sound-scanner

.\scripts\fetch-native.ps1

## once more
go build -o (Join-Path $PWD.Path 'bin/win-sound-scanner.exe') ./cmd/win-sound-scanner
```

Alternative: run `.\scripts\build.ps1` (or `.\scripts\build.ps1 -m ""`).

## Run
```powershell
.\bin\win-sound-scanner.exe
```

## Windows Service
The executable supports both modes:
- Console mode (interactive): `.\bin\win-sound-scanner.exe`
- Service mode (managed by SCM): install/start/stop via commands below

Service commands (run in elevated PowerShell):
```powershell
.\bin\win-sound-scanner.exe install
.\bin\win-sound-scanner.exe start
.\bin\win-sound-scanner.exe stop
.\bin\win-sound-scanner.exe restart
.\bin\win-sound-scanner.exe uninstall
```

To store RabbitMQ settings as service environment variables, set them before `install`:
```powershell
$Env:WIN_SOUND_ENQUEUER = "rabbitmq"
$Env:WIN_SOUND_RABBITMQ_HOST = "localhost"
$Env:WIN_SOUND_RABBITMQ_PORT = "5672"

.\bin\win-sound-scanner.exe install
```
Only currently defined `WIN_SOUND_*` variables are written into the service config.
If you change service env vars later, run `stop`, `uninstall`, `install`, `start`.

## RabbitMQ Mode (scanner)
By default, scanner uses the RabbitMQ-enqueuer (`WIN_SOUND_ENQUEUER=rabbitmq`).
To disable request publishing to RabbitMQ, set it to `empty`:
```powershell
$Env:WIN_SOUND_ENQUEUER = "empty"

# Optional overrides (defaults shown)
$Env:WIN_SOUND_RABBITMQ_HOST = "localhost"
$Env:WIN_SOUND_RABBITMQ_PORT = "5672"
$Env:WIN_SOUND_RABBITMQ_VHOST = "/"
$Env:WIN_SOUND_RABBITMQ_USER = "guest"
$Env:WIN_SOUND_RABBITMQ_PASSWORD = "guest"
$Env:WIN_SOUND_RABBITMQ_EXCHANGE = "sdr_exchange"
$Env:WIN_SOUND_RABBITMQ_QUEUE = "sdr_queue"
$Env:WIN_SOUND_RABBITMQ_ROUTING_KEY = "sdr_bind"
```

## Debug
Compile with -gcflags=all="-N -l" to disable optimizations and inlining, then run with a debugger
```powershell
go build -gcflags "all=-N -l" -o (Join-Path $PWD.Path 'bin/win-sound-scanner.exe') ./cmd/win-sound-scanner
dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/win-sound-scanner.exe
```
Then use remote debugging in your IDE (e.g. GoLand) to connect to localhost:2345

## External module
- github.com/collect-sound-devices/sound-win-scanner/v4 (pkg/soundlibwrap): cgo wrapper around SoundAgentApi, see [soundlibwrap documentation](https://pkg.go.dev/github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap)

## License

This project is licensed under the terms of the [MIT License](LICENSE).

## Contact

Eduard Danziger

Email: [edanziger@gmx.de](mailto:edanziger@gmx.de)
