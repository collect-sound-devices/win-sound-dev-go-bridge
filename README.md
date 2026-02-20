# win-sound-scanner (Windows Sound Scanner, WinSoundScanner)

WinSoundScanner detects default audio endpoint devices under Windows.
Besides, it handles thew following audio notifications: sound device changes and sound volume adjustments.

The WinSoundScanner registers audio device information on a backend server via REST API, pushing the respective request-messages to RabbitMQ channel.
The separate RMQ To REST API Forwarder .NET Windows Service fetches these messages and forwards them to the REST API calls.
The respective backend, Audio Device Repository Server (ASP.Net Core), resides in [audio-device-repo-server](https://github.com/collect-sound-devices/audio-device-repo-server/) with a React / TypeScript frontend [list-audio-react-app](https://github.com/collect-sound-devices/list-audio-react-app/), see [Primary Web Client](https://list-audio-react-app.vercel.app).
![primaryWebClient screenshot](202509011555ReactRepoApp.jpg)

Technically, WinSoundScanner is a Go(Golang)+CGO bridge to the C++ `SoundAgentApi.dll` (monitoring/querying Windows Dll) and event redirector.
It can run as a console application or as a Windows Service.

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
Service logs are written to:
`%ProgramData%\WinSoundScanner\service.log`

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
```

Optional overrides with default values:
```powershell
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
