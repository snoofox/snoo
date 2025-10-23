package debug

import (
	"fmt"
	"os"
	"time"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
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
