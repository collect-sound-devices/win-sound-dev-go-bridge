package logging

import (
	"log"
	"os"
	"strings"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

// Logger is the minimal interface needed for logging in this project.
type Logger interface {
	Printf(format string, v ...interface{})
}

// NewAppLogger builds the default logger used by the scanner app.
func NewAppLogger() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// NewPlainLogger builds a logger without any prefix/flags, useful for bridging messages that already include timestamps.
func NewPlainLogger() *log.Logger {
	return log.New(os.Stdout, "", 0)
}

func PrintInfo(logger Logger, format string, v ...interface{}) {
	if logger == nil {
		return
	}
	logger.Printf("[info] "+format, v...)
}

func PrintError(logger Logger, format string, v ...interface{}) {
	if logger == nil {
		return
	}
	logger.Printf("[error] "+format, v...)
}

// AttachSoundlibwrapBridge forwards soundlibwrap log messages into the provided logger.
// The logger should typically have no flags/prefix so the embedded timestamp is preserved.
func AttachSoundlibwrapBridge(logger Logger, prefix string) {
	if logger == nil {
		return
	}
	if prefix == "" {
		prefix = "cpp backend"
	}

	soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
		lvl := strings.ToLower(level)
		levelTag := "info"
		switch lvl {
		case "trace", "debug":
			levelTag = "debug"
		case "warn", "warning":
			levelTag = "warn"
		case "error", "critical":
			levelTag = "error"
		}

		logger.Printf("%s [%s %s] %s", timestamp, prefix, levelTag, content)
	})
}
