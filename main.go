package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/cliapp"
)

var (
	modOle32           = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize = modOle32.NewProc("CoUninitialize")
)

//goland:noinspection ALL
const (
	COINIT_APARTMENTTHREADED = 0x2 // Single-threaded apartment
	COINIT_MULTITHREADED     = 0x0 // Multithreaded apartment
)

// suppress unused
var _ = COINIT_APARTMENTTHREADED
var _ = COINIT_MULTITHREADED

func CoInitializeEx(coInit uintptr) error {
	ret, _, _ := procCoInitializeEx.Call(0, coInit)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func CoUninitialize() {
	procCoUninitialize.Call() // best‑effort cleanup; failure is ignored
}

func main() {
	if err := CoInitializeEx(COINIT_MULTITHREADED); err != nil {
		log.Fatalf("COM initialization failed: %v", err)
	}
	defer CoUninitialize()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cliapp.Run(ctx); err != nil {
		log.Fatalf("exit with error: %v", err)
	}
}
