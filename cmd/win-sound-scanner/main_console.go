package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/scannerapp"
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
	procCoUninitialize.Call() // best-effort cleanup; failure is ignored
}

func runScanner(ctx context.Context) error {
	if err := CoInitializeEx(COINIT_MULTITHREADED); err != nil {
		return fmt.Errorf("COM initialization failed: %w", err)
	}
	defer CoUninitialize()

	if err := scannerapp.Run(ctx); err != nil {
		return fmt.Errorf("scanner run failed: %w", err)
	}
	return nil
}

func runConsole() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runScanner(ctx)
}
