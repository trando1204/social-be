package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"socialat/be/webserver"
	"socialat/be/webserver/service"

	"github.com/decred/slog"
	"github.com/jrick/logrotate/rotator"
)

var socialatLog = "socialat.log"

var (
	logRotator *rotator.Rotator
	backendLog = slog.NewBackend(logWriter{socialatLog})
	Log        = backendLog.Logger("SOCIALAT")
)

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct {
	loggerID string
}

// Write writes the data in p to standard out and the log rotator.
func (l logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return logRotator.Write(p)
}

func GetDBLogger() *log.Logger {
	logger := log.New(os.Stdout, "\r\n[DB] ", log.LstdFlags)
	logger.SetOutput(logWriter{})
	return logger
}

func initLog() {
	webserver.UseLogger(Log)
	service.UseLogger(Log)
}

func SetLogLevel(logLevel string) {
	level, _ := slog.LevelFromString(logLevel)
	Log.SetLevel(level)
}

func InitLogRotator(logDir string) error {
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		return err
	}

	r, err := rotator.New(filepath.Join(logDir, socialatLog), 32*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		return err
	}

	logRotator = r
	initLog()
	return nil
}
