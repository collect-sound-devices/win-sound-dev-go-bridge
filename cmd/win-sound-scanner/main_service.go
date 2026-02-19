package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kardianos/service"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/scannerapp"
)

const (
	serviceName        = "WinSoundScanner"
	serviceDisplayName = "Win Sound Scanner"
	serviceDescription = "Collects Windows default sound devices and publishes events."
	serviceLogFileName = "service.log"
)

var serviceEnvKeys = []string{
	scannerapp.EnvWinSoundEnqueuer,
	scannerapp.EnvWinSoundRabbitMQHost,
	scannerapp.EnvWinSoundRabbitMQPort,
	scannerapp.EnvWinSoundRabbitMQVHost,
	scannerapp.EnvWinSoundRabbitMQUser,
	scannerapp.EnvWinSoundRabbitMQPassword,
	scannerapp.EnvWinSoundRabbitMQExchange,
	scannerapp.EnvWinSoundRabbitMQQueue,
	scannerapp.EnvWinSoundRabbitMQRoutingKey,
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
