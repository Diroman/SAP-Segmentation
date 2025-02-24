package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"sap_segmentation/config"
	"sap_segmentation/model"
)

func main() {
	ctx := context.Background()

	// read config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(errors.Wrap(err, "failed to load config: %v"))
	}

	// clean old logs before creating new file
	cleanupOldLogs(cfg)

	// setup logger
	logger := setupLogger()

	segmentation, err := model.NewSegmentation(&cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to create segmentation service")
	}

	if err = segmentation.Run(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to run segmentation")
	}
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()

	if err := os.MkdirAll("./log", 0755); err != nil {
		logrus.WithError(err).Fatal("failed to create log directory")
	}

	logFile, err := os.OpenFile("./log/segmentation_import.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.WithError(err).Fatal("failed to open log file")
	}

	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logLevel := "info"
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warnf("invalid log level '%s', defaulting to info", logLevel)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

func cleanupOldLogs(cfg config.Config) {
	files, err := filepath.Glob("/log/*.log")
	if err != nil {
		log.Printf("Error reading log directory: %v", err)
		return
	}

	maxAge := time.Duration(cfg.LogCleanupMaxAge) * 24 * time.Hour
	now := time.Now()

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			continue
		}
		if now.Sub(fi.ModTime()) > maxAge {
			if err = os.Remove(file); err != nil {
				log.Printf("Error removing old log file %s: %v", file, err)
			}
		}
	}
}
