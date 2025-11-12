package config

import "time"

// Config holds the configuration for the collector.
type Config struct {
	Namespace           string
	OBTenant            string
	Interval            time.Duration
	DataPath            string
	QueueSize           int
	WorkerNum           int
	CompactionThreshold int
	SqlAuditLimit            int
	SlowSqlThresholdMilliSeconds int
}
