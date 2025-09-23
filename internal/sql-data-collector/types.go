package sqldatacollector

// SQLAudit represents a single row of aggregated data from gv$ob_sql_audit.
type SQLAudit struct {
	SvrIP        string `db:"svr_ip"`
	TenantID     int64  `db:"tenant_id"`
	TenantName   string `db:"tenant_name"`
	UserName     string `db:"user_name"`
	DatabaseName string `db:"database_name"`
	SQLID        string `db:"sql_id"`
	QuerySQL     string `db:"query_sql"`
	PlanID       int64  `db:"plan_id"`
	AffectedRows int64  `db:"affected_rows"`
	ReturnRows   int64  `db:"return_rows"`
	RetCode      int64  `db:"ret_code"`
	RequestID    uint64 `db:"request_id"`
	RequestTime  int64  `db:"request_time"`
	ElapsedTime  int64  `db:"elapsed_time"`
	ExecuteTime  int64  `db:"execute_time"`
	QueueTime    int64  `db:"queue_time"`
}

// Tenant represents a tenant with its ID.
type Tenant struct {
	ID int64 `db:"tenant_id"`
}
