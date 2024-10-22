package logging

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()

	// Ensure the logs directory exists
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create a log file if it doesn't exist, and set the file as the log output
	file, err := os.OpenFile(logDir+"/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Set up multi-writer to write to both the log file and stdout (console)
	multiWriter := io.MultiWriter(file, os.Stdout)
	Log.SetOutput(multiWriter)

	// Set log format to JSON for structured logging
	Log.SetFormatter(&logrus.JSONFormatter{})

	// Set log level (can be changed to DebugLevel for more verbosity)
	Log.SetLevel(logrus.InfoLevel)
}
