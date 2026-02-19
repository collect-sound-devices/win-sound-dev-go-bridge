package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/kardianos/service"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/scannerapp"
)

var (
	modOle32           = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeEx = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize = modOle32.NewProc("CoUninitialize")
)

const (
	serviceName        = "WinSoundScanner"
	serviceDisplayName = "Win Sound Scanner"
	serviceDescription = "Collects Windows default sound devices and publishes events."
	serviceLogFileName = "service.log"
)

var serviceEnvKeys = []string{
	"WIN_SOUND_ENQUEUER",
	"WIN_SOUND_RABBITMQ_HOST",
	"WIN_SOUND_RABBITMQ_PORT",
	"WIN_SOUND_RABBITMQ_VHOST",
	"WIN_SOUND_RABBITMQ_USER",
	"WIN_SOUND_RABBITMQ_PASSWORD",
	"WIN_SOUND_RABBITMQ_EXCHANGE",
	"WIN_SOUND_RABBITMQ_QUEUE",
	"WIN_SOUND_RABBITMQ_ROUTING_KEY",
}

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

type scannerProgram struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func (p *scannerProgram) Start(_ service.Service) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.done != nil {
		return nil
	}

	if err := configureServiceFileLogging(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	p.cancel = cancel
	p.done = done

	go func() {
		defer close(done)
		if err := runScanner(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("scanner failed: %v", err)
			os.Exit(1)
		}
	}()

	return nil
}

func (p *scannerProgram) Stop(_ service.Service) error {
	p.mu.Lock()
	cancel := p.cancel
	done := p.done
	p.cancel = nil
	p.done = nil
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	return nil
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

func newService() (service.Service, error) {
	envVars := collectServiceEnvVars()
	if len(envVars) == 0 {
		envVars = nil
	}

	cfg := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		Option: service.KeyValue{
			"StartType": "automatic",
			"OnFailure": "restart",
		},
		EnvVars: envVars,
	}

	return service.New(&scannerProgram{}, cfg)
}

func collectServiceEnvVars() map[string]string {
	envVars := make(map[string]string, len(serviceEnvKeys))
	for _, key := range serviceEnvKeys {
		if value, ok := os.LookupEnv(key); ok {
			envVars[key] = value
		}
	}
	return envVars
}

func isServiceCommand(cmd string) bool {
	switch cmd {
	case "install", "uninstall", "start", "stop", "restart":
		return true
	default:
		return false
	}
}

func programDataDir() (string, error) {
	if v, ok := os.LookupEnv("ProgramData"); ok && strings.TrimSpace(v) != "" {
		return v, nil
	}
	if v, ok := os.LookupEnv("ALLUSERSPROFILE"); ok && strings.TrimSpace(v) != "" {
		return v, nil
	}
	return "", errors.New("ProgramData is not available in environment")
}

func configureServiceFileLogging() error {
	baseDir, err := programDataDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(baseDir, serviceName)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create service log directory: %w", err)
	}

	logPath := filepath.Join(logDir, serviceLogFileName)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open service log file: %w", err)
	}

	log.SetOutput(logFile)
	os.Stdout = logFile
	os.Stderr = logFile
	return nil
}

func main() {
	if len(os.Args) > 1 {
		cmd := strings.ToLower(strings.TrimSpace(os.Args[1]))
		if !isServiceCommand(cmd) {
			log.Fatalf("unsupported command %q (supported: install, uninstall, start, stop, restart)", cmd)
		}

		svc, err := newService()
		if err != nil {
			log.Fatalf("service initialization failed: %v", err)
		}
		if err := service.Control(svc, cmd); err != nil {
			log.Fatalf("service command %q failed: %v", cmd, err)
		}
		return
	}

	if service.Interactive() {
		if err := runConsole(); err != nil {
			log.Fatalf("exit with error: %v", err)
		}
		return
	}

	svc, err := newService()
	if err != nil {
		log.Fatalf("service initialization failed: %v", err)
	}
	if err := svc.Run(); err != nil {
		log.Fatalf("service run failed: %v", err)
	}
}
