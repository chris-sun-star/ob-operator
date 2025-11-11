package parquet

const (
	CompactionThreshold = 120
)

const (
	SmallFilePattern     = "[0-9]*.parquet"
	CompactedFilePattern = "compacted-*.parquet"
	FileTimeFormat       = "2006-01-02-15-04-05"
)
