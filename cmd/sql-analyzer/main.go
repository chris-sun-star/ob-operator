package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	logger "github.com/sirupsen/logrus"

	webserver "github.com/oceanbase/ob-operator/internal/server"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/collector"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/router"
	"github.com/oceanbase/ob-operator/pkg/log"
)

const (
	CompactionThreshold = 120
	PlanWorkerCount     = 4
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "log/sql-analyzer.log"
	}
	log.InitLogger(
		log.LoggerConfig{
			Level:      logLevel,
			Filename:   logFile,
			MaxSize:    256,
			MaxAge:     7,
			MaxBackups: 5,
			LocalTime:  true,
			Compress:   true,
		},
	)
}

func startHttpServer() {
	httpServer := webserver.NewHTTPServer()
	router.Register(httpServer.Router)
	logger.Info("Successfully registered router")
	err := httpServer.Run()
	if err != nil {
		logger.WithError(err).Errorln("Start server failed")
		os.Exit(1)
	}
}

func main() {
	// Read configuration from environment variables.
	namespace := os.Getenv("NAMESPACE")
	obtenant := os.Getenv("OBTENANT")
	dataPath := os.Getenv("DATA_PATH")

	if namespace == "" || obtenant == "" {
		logger.Fatalf("NAMESPACE, OBTENANT environment variables must be set.")
	}
	if dataPath == "" {
		dataPath = "."
	}
	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	// Configure collection interval
	intervalSeconds := 30
	intervalStr := os.Getenv("COLLECTION_INTERVAL_SECONDS")
	if intervalStr != "" {
		if val, err := strconv.Atoi(intervalStr); err == nil && val > 0 {
			intervalSeconds = val
		} else {
			logger.Printf("Invalid COLLECTION_INTERVAL_SECONDS value '%s', using default of 30 seconds.", intervalStr)
		}
	}

	config := &config.Config{
		Namespace: namespace,
		OBTenant:  obtenant,
		Interval:  time.Duration(intervalSeconds) * time.Second,
		DataPath:  dataPath,
		// config via environment variable
		QueueSize: 100,
		WorkerNum: 4,
	}

	collector := collector.NewCollector(ctx, config)
	collector.Start()

	startHttpServer()
}
