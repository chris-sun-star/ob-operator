package sqldatacollector

// SQLAudit represents an aggregated row of data from gv$ob_sql_audit.
type SQLAudit struct {
	// Grouping Keys
	SvrIP             string `db:"svr_ip"`
	TenantId          uint64 `db:"tenant_id"`
	TenantName        string `db:"tenant_name"`
	UserId            int64  `db:"user_id"`
	UserName          string `db:"user_name"`
	DbId              uint64 `db:"db_id"`
	DBName            string `db:"db_name"`
	SqlId             string `db:"sql_id"`
	QuerySql          string `db:"query_sql"`
	PlanId            int64  `db:"plan_id"`
	ClientIp          string `db:"client_ip"`
	Event             string `db:"event"`
	PlanType          int64  `db:"plan_type"`
	ConsistencyLevel  int64  `db:"consistency_level"`

	// Aggregated Values
	Executions        int64 `db:"executions"`
	MinRequestTime    int64 `db:"min_request_time"`
	MaxRequestTime    int64 `db:"max_request_time"`

	ElapsedTimeSum int64 `db:"elapsed_time_sum"`
	ElapsedTimeMax int64 `db:"elapsed_time_max"`
	ElapsedTimeMin int64 `db:"elapsed_time_min"`

	ExecuteTimeSum int64 `db:"execute_time_sum"`
	ExecuteTimeMax int64 `db:"execute_time_max"`
	ExecuteTimeMin int64 `db:"execute_time_min"`

	QueueTimeSum int64 `db:"queue_time_sum"`
	QueueTimeMax int64 `db:"queue_time_max"`
	QueueTimeMin int64 `db:"queue_time_min"`

	GetPlanTimeSum int64 `db:"get_plan_time_sum"`
	GetPlanTimeMax int64 `db:"get_plan_time_max"`
	GetPlanTimeMin int64 `db:"get_plan_time_min"`

	AffectedRowsSum int64 `db:"affected_rows_sum"`
	AffectedRowsMax int64 `db:"affected_rows_max"`
	AffectedRowsMin int64 `db:"affected_rows_min"`

	ReturnRowsSum int64 `db:"return_rows_sum"`
	ReturnRowsMax int64 `db:"return_rows_max"`
	ReturnRowsMin int64 `db:"return_rows_min"`

	PartitionCountSum int64 `db:"partition_count_sum"`
	PartitionCountMax int64 `db:"partition_count_max"`
	PartitionCountMin int64 `db:"partition_count_min"`

	RetryCountSum int64 `db:"retry_count_sum"`
	RetryCountMax int64 `db:"retry_count_max"`
	RetryCountMin int64 `db:"retry_count_min"`

	DiskReadsSum int64 `db:"disk_reads_sum"`
	DiskReadsMax int64 `db:"disk_reads_max"`
	DiskReadsMin int64 `db:"disk_reads_min"`

	RpcCountSum int64 `db:"rpc_count_sum"`
	RpcCountMax int64 `db:"rpc_count_max"`
	RpcCountMin int64 `db:"rpc_count_min"`

	MemstoreReadRowCountSum int64 `db:"memstore_read_row_count_sum"`
	MemstoreReadRowCountMax int64 `db:"memstore_read_row_count_max"`
	MemstoreReadRowCountMin int64 `db:"memstore_read_row_count_min"`

	SSStoreReadRowCountSum int64 `db:"ssstore_read_row_count_sum"`
	SSStoreReadRowCountMax int64 `db:"ssstore_read_row_count_max"`
	SSStoreReadRowCountMin int64 `db:"ssstore_read_row_count_min"`

	RequestMemoryUsedSum int64 `db:"request_memory_used_sum"`
	RequestMemoryUsedMax int64 `db:"request_memory_used_max"`
	RequestMemoryUsedMin int64 `db:"request_memory_used_min"`

	FailCountSum int64 `db:"fail_count_sum"`

	RetCode4012CountSum int64 `db:"ret_code_4012_count_sum"`
	RetCode4013CountSum int64 `db:"ret_code_4013_count_sum"`
	RetCode5001CountSum int64 `db:"ret_code_5001_count_sum"`
	RetCode5024CountSum int64 `db:"ret_code_5024_count_sum"`
	RetCode5167CountSum int64 `db:"ret_code_5167_count_sum"`
	RetCode5217CountSum int64 `db:"ret_code_5217_count_sum"`
	RetCode6002CountSum int64 `db:"ret_code_6002_count_sum"`

	Event0WaitTimeSum int64 `db:"event_0_wait_time_sum"`
	Event1WaitTimeSum int64 `db:"event_1_wait_time_sum"`
	Event2WaitTimeSum int64 `db:"event_2_wait_time_sum"`
	Event3WaitTimeSum int64 `db:"event_3_wait_time_sum"`
}

// Tenant represents a tenant with its ID.
type Tenant struct {
	ID int64 `db:"tenant_id"`
}