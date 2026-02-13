package rabbitmq

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost                    = "localhost"
	defaultPort                    = 5672
	defaultVHost                   = "/"
	defaultUser                    = "guest"
	defaultPassword                = "guest"
	defaultExchangeName            = "sdr_exchange"
	defaultQueueName               = "sdr_queue"
	defaultRoutingKey              = "sdr_bind"
	defaultConnectionThreshold     = 20 * time.Second
	defaultMaxReconnectionAttempts = 8
	defaultInitialReconnectDelay   = 2 * time.Second
	defaultMaxReconnectDelay       = 30 * time.Second
	defaultPublishConfirmTimeout   = 10 * time.Second
)

// Config defines RabbitMQ connection, topology, and retry settings.
type Config struct {
	Host                    string
	Port                    int
	VHost                   string
	User                    string
	Password                string
	ExchangeName            string
	QueueName               string
	RoutingKey              string
	ConnectionThreshold     time.Duration
	MaxReconnectionAttempts int
	InitialReconnectDelay   time.Duration
	MaxReconnectDelay       time.Duration
	PublishConfirmTimeout   time.Duration
}

func DefaultConfig() Config {
	return Config{
		Host:                    defaultHost,
		Port:                    defaultPort,
		VHost:                   defaultVHost,
		User:                    defaultUser,
		Password:                defaultPassword,
		ExchangeName:            defaultExchangeName,
		QueueName:               defaultQueueName,
		RoutingKey:              defaultRoutingKey,
		ConnectionThreshold:     defaultConnectionThreshold,
		MaxReconnectionAttempts: defaultMaxReconnectionAttempts,
		InitialReconnectDelay:   defaultInitialReconnectDelay,
		MaxReconnectDelay:       defaultMaxReconnectDelay,
		PublishConfirmTimeout:   defaultPublishConfirmTimeout,
	}
}

func (c Config) withDefaults() Config {
	d := DefaultConfig()
	if strings.TrimSpace(c.Host) == "" {
		c.Host = d.Host
	}
	if c.Port <= 0 {
		c.Port = d.Port
	}
	if c.VHost == "" {
		c.VHost = d.VHost
	}
	if c.User == "" {
		c.User = d.User
	}
	if c.Password == "" {
		c.Password = d.Password
	}
	if c.ExchangeName == "" {
		c.ExchangeName = d.ExchangeName
	}
	if c.QueueName == "" {
		c.QueueName = d.QueueName
	}
	if c.RoutingKey == "" {
		c.RoutingKey = d.RoutingKey
	}
	if c.ConnectionThreshold <= 0 {
		c.ConnectionThreshold = d.ConnectionThreshold
	}
	if c.MaxReconnectionAttempts <= 0 {
		c.MaxReconnectionAttempts = d.MaxReconnectionAttempts
	}
	if c.InitialReconnectDelay <= 0 {
		c.InitialReconnectDelay = d.InitialReconnectDelay
	}
	if c.MaxReconnectDelay <= 0 {
		c.MaxReconnectDelay = d.MaxReconnectDelay
	}
	if c.PublishConfirmTimeout <= 0 {
		c.PublishConfirmTimeout = d.PublishConfirmTimeout
	}
	if c.MaxReconnectDelay < c.InitialReconnectDelay {
		c.MaxReconnectDelay = c.InitialReconnectDelay
	}
	return c
}

// LoadConfigFromEnv loads rabbit configuration from environment variables.
// Empty values are replaced by defaults.
func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_HOST")); v != "" {
		cfg.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_PORT")); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_PORT %q: %w", v, err)
		}
		cfg.Port = port
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_VHOST"); v != "" {
		cfg.VHost = v
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_USER"); v != "" {
		cfg.User = v
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_EXCHANGE"); v != "" {
		cfg.ExchangeName = v
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_QUEUE"); v != "" {
		cfg.QueueName = v
	}
	if v := os.Getenv("WIN_SOUND_RABBITMQ_ROUTING_KEY"); v != "" {
		cfg.RoutingKey = v
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_CONNECTION_THRESHOLD_SEC")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_CONNECTION_THRESHOLD_SEC %q: %w", v, err)
		}
		cfg.ConnectionThreshold = time.Duration(n) * time.Second
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_MAX_RECONNECT_ATTEMPTS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_MAX_RECONNECT_ATTEMPTS %q: %w", v, err)
		}
		cfg.MaxReconnectionAttempts = n
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_INITIAL_RECONNECT_DELAY_MS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_INITIAL_RECONNECT_DELAY_MS %q: %w", v, err)
		}
		cfg.InitialReconnectDelay = time.Duration(n) * time.Millisecond
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_MAX_RECONNECT_DELAY_MS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_MAX_RECONNECT_DELAY_MS %q: %w", v, err)
		}
		cfg.MaxReconnectDelay = time.Duration(n) * time.Millisecond
	}
	if v := strings.TrimSpace(os.Getenv("WIN_SOUND_RABBITMQ_PUBLISH_CONFIRM_TIMEOUT_MS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WIN_SOUND_RABBITMQ_PUBLISH_CONFIRM_TIMEOUT_MS %q: %w", v, err)
		}
		cfg.PublishConfirmTimeout = time.Duration(n) * time.Millisecond
	}

	return cfg.withDefaults(), nil
}
