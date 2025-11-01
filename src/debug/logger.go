package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

func init() {
	logPath, err := getLogPath()
	if err != nil {
		return
	}

	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
}

func getLogPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	snooDir := filepath.Join(homeDir, ".snoo")
	if err := os.MkdirAll(snooDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(snooDir, "debug.log"), nil
}

func Log(format string, args ...interface{}) {
	if logFile == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
