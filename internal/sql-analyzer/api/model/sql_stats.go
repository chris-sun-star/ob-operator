/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package model

// QuerySqlStatsRequest defines the request body for querying SQL statistics.
type QuerySqlStatsRequest struct {
	StartTime       string   `json:"startTime" binding:"required" example:"2025-11-17T10:00:00Z"`
	EndTime         string   `json:"endTime" binding:"required" example:"2025-11-17T11:00:00Z"`
	UserName        string   `json:"userName,omitempty" example:"user1"`
	DatabaseName    string   `json:"databaseName,omitempty" example:"db1"`
	FilterInnerSql  bool     `json:"filterInnerSql,omitempty"`
	QuerySqlKeyword string   `json:"querySqlKeyword,omitempty" example:"SELECT"`
	SortByColumn    string   `json:"sortByColumn,omitempty" example:"request_time"`
	SortOrder       string   `json:"sortOrder,omitempty" example:"DESC"`
	OutputColumns   []string `json:"outputColumns,omitempty" example:"query_sql,request_time,affected_rows"`
}

// SqlAuditItem defines a single item in the SQL audit statistics response.
// It mirrors the main SqlAudit struct but uses camelCase JSON tags for the API.
type SqlStatsItem struct {
	SvrIP      string `json:"svrIp"`
	SvrPort    int64  `json:"svrPort"`
	TenantId   uint64 `json:"tenantId"`
	TenantName string `json:"tenantName"`
	UserId     int64  `json:"userId"`
	UserName   string `json:"userName"`
	DbId       uint64 `json:"dbId"`
	DBName     string `json:"dbName"`
	SqlId      string `json:"sqlId"`
	PlanId     int64  `json:"planId"`

	QuerySql          string `json:"querySql"`
	ClientIp          string `json:"clientIp"`
	Event             string `json:"event"`
	FormatSqlId       string `json:"formatSqlId"`
	EffectiveTenantId uint64 `json:"effectiveTenantId"`
	TraceId           string `json:"traceId"`
	Sid               uint64 `json:"sid"`
	UserClientIp      string `json:"userClientIp"`
	TxId              string `json:"txId"`
	SubPlanCount      int64  `json:"subPlanCount"`
	LastFailInfo      int64  `json:"lastFailInfo"`
	CauseType         int64  `json:"causeType"`

	Executions     int64  `json:"executions"`
	MinRequestTime int64  `json:"minRequestTime"`
	MaxRequestTime int64  `json:"maxRequestTime"`
	MaxRequestId   uint64 `json:"maxRequestId"`
	MinRequestId   uint64 `json:"minRequestId"`

	ElapsedTimeSum int64 `json:"elapsedTimeSum"`
	ElapsedTimeMax int64 `json:"elapsedTimeMax"`
	ElapsedTimeMin int64 `json:"elapsedTimeMin"`

	ExecuteTimeSum int64 `json:"executeTimeSum"`
	ExecuteTimeMax int64 `json:"executeTimeMax"`
	ExecuteTimeMin int64 `json:"executeTimeMin"`

	QueueTimeSum int64 `json:"queueTimeSum"`
	QueueTimeMax int64 `json:"queueTimeMax"`
	QueueTimeMin int64 `json:"queueTimeMin"`

	GetPlanTimeSum int64 `json:"getPlanTimeSum"`
	GetPlanTimeMax int64 `json:"getPlanTimeMax"`
	GetPlanTimeMin int64 `json:"getPlanTimeMin"`

	AffectedRowsSum int64 `json:"affectedRowsSum"`
	AffectedRowsMax int64 `json:"affectedRowsMax"`
	AffectedRowsMin int64 `json:"affectedRowsMin"`

	ReturnRowsSum int64 `json:"returnRowsSum"`
	ReturnRowsMax int64 `json:"returnRowsMax"`
	ReturnRowsMin int64 `json:"returnRowsMin"`

	PartitionCountSum int64 `json:"partitionCountSum"`
	PartitionCountMax int64 `json:"partitionCountMax"`
	PartitionCountMin int64 `json:"partitionCountMin"`

	RetryCountSum int64 `json:"retryCountSum"`
	RetryCountMax int64 `json:"retryCountMax"`
	RetryCountMin int64 `json:"retryCountMin"`

	DiskReadsSum int64 `json:"diskReadsSum"`
	DiskReadsMax int64 `json:"diskReadsMax"`
	DiskReadsMin int64 `json:"diskReadsMin"`

	RpcCountSum int64 `json:"rpcCountSum"`
	RpcCountMax int64 `json:"rpcCountMax"`
	RpcCountMin int64 `json:"rpcCountMin"`

	MemstoreReadRowCountSum int64 `json:"memstoreReadRowCountSum"`
	MemstoreReadRowCountMax int64 `json:"memstoreReadRowCountMax"`
	MemstoreReadRowCountMin int64 `json:"memstoreReadRowCountMin"`

	SSStoreReadRowCountSum int64 `json:"ssstoreReadRowCountSum"`
	SSStoreReadRowCountMax int64 `json:"ssstoreReadRowCountMax"`
	SSStoreReadRowCountMin int64 `json:"ssstoreReadRowCountMin"`

	RequestMemoryUsedSum int64 `json:"requestMemoryUsedSum"`
	RequestMemoryUsedMax int64 `json:"requestMemoryUsedMax"`
	RequestMemoryUsedMin int64 `json:"requestMemoryUsedMin"`

	WaitTimeMicroSum int64 `json:"waitTimeMicroSum"`
	WaitTimeMicroMax int64 `json:"waitTimeMicroMax"`
	WaitTimeMicroMin int64 `json:"waitTimeMicroMin"`

	TotalWaitTimeMicroSum int64 `json:"totalWaitTimeMicroSum"`
	TotalWaitTimeMicroMax int64 `json:"totalWaitTimeMicroMax"`
	TotalWaitTimeMicroMin int64 `json:"totalWaitTimeMicroMin"`

	NetTimeSum int64 `json:"netTimeSum"`
	NetTimeMax int64 `json:"netTimeMax"`
	NetTimeMin int64 `json:"netTimeMin"`

	NetWaitTimeSum int64 `json:"netWaitTimeSum"`
	NetWaitTimeMax int64 `json:"netWaitTimeMax"`
	NetWaitTimeMin int64 `json:"netWaitTimeMin"`

	DecodeTimeSum int64 `json:"decodeTimeSum"`
	DecodeTimeMax int64 `json:"decodeTimeMax"`
	DecodeTimeMin int64 `json:"decodeTimeMin"`

	ApplicationWaitTimeSum int64 `json:"applicationWaitTimeSum"`
	ApplicationWaitTimeMax int64 `json:"applicationWaitTimeMax"`
	ApplicationWaitTimeMin int64 `json:"applicationWaitTimeMin"`

	ConcurrencyWaitTimeSum int64 `json:"concurrencyWaitTimeSum"`
	ConcurrencyWaitTimeMax int64 `json:"concurrencyWaitTimeMax"`
	ConcurrencyWaitTimeMin int64 `json:"concurrencyWaitTimeMin"`

	UserIoWaitTimeSum int64 `json:"userIoWaitTimeSum"`
	UserIoWaitTimeMax int64 `json:"userIoWaitTimeMax"`
	UserIoWaitTimeMin int64 `json:"userIoWaitTimeMin"`

	ScheduleTimeSum int64 `json:"scheduleTimeSum"`
	ScheduleTimeMax int64 `json:"scheduleTimeMax"`
	ScheduleTimeMin int64 `json:"scheduleTimeMin"`

	RowCacheHitSum int64 `json:"rowCacheHitSum"`
	RowCacheHitMax int64 `json:"rowCacheHitMax"`
	RowCacheHitMin int64 `json:"rowCacheHitMin"`

	BloomFilterCacheHitSum int64 `json:"bloomFilterCacheHitSum"`
	BloomFilterCacheHitMax int64 `json:"bloomFilterCacheHitMax"`
	BloomFilterCacheHitMin int64 `json:"bloomFilterCacheHitMin"`

	BlockCacheHitSum int64 `json:"blockCacheHitSum"`
	BlockCacheHitMax int64 `json:"blockCacheHitMax"`
	BlockCacheHitMin int64 `json:"blockCacheHitMin"`

	IndexBlockCacheHitSum int64 `json:"indexBlockCacheHitSum"`
	IndexBlockCacheHitMax int64 `json:"indexBlockCacheHitMax"`
	IndexBlockCacheHitMin int64 `json:"indexBlockCacheHitMin"`

	ExpectedWorkerCountSum int64 `json:"expectedWorkerCountSum"`
	ExpectedWorkerCountMax int64 `json:"expectedWorkerCountMax"`
	ExpectedWorkerCountMin int64 `json:"expectedWorkerCountMin"`

	UsedWorkerCountSum int64 `json:"usedWorkerCountSum"`
	UsedWorkerCountMax int64 `json:"usedWorkerCountMax"`
	UsedWorkerCountMin int64 `json:"usedWorkerCountMin"`

	TableScanSum int64 `json:"tableScanSum"`
	TableScanMax int64 `json:"tableScanMax"`
	TableScanMin int64 `json:"tableScanMin"`

	ConsistencyLevelStrongCount int64 `json:"consistencyLevelStrongCount"`
	ConsistencyLevelWeakCount   int64 `json:"consistencyLevelWeakCount"`

	CpuTimeSum int64 `json:"cpuTimeSum"`
	CpuTimeMax int64 `json:"cpuTimeMax"`
	CpuTimeMin int64 `json:"cpuTimeMin"`

	FailCountSum int64 `json:"failCountSum"`

	RetCode4012CountSum int64 `json:"retCode4012CountSum"`
	RetCode4013CountSum int64 `json:"retCode4013CountSum"`
	RetCode5001CountSum int64 `json:"retCode5001CountSum"`
	RetCode5024CountSum int64 `json:"retCode5024CountSum"`
	RetCode5167CountSum int64 `json:"retCode5167CountSum"`
	RetCode5217CountSum int64 `json:"retCode5217CountSum"`
	RetCode6002CountSum int64 `json:"retCode6002CountSum"`

	Event0WaitTimeSum int64 `json:"event0WaitTimeSum"`
	Event1WaitTimeSum int64 `json:"event1WaitTimeSum"`
	Event2WaitTimeSum int64 `json:"event2WaitTimeSum"`
	Event3WaitTimeSum int64 `json:"event3WaitTimeSum"`

	PlanTypeLocalCount       int64 `json:"planTypeLocalCount"`
	PlanTypeRemoteCount      int64 `json:"planTypeRemoteCount"`
	PlanTypeDistributedCount int64 `json:"planTypeDistributedCount"`
	InnerSqlCount            int64 `json:"innerSqlCount"`
	MissPlanCount            int64 `json:"missPlanCount"`
	ExecutorRpcCount         int64 `json:"executorRpcCount"`
}

// SqlStatsResponse defines the overall structure of the SQL statistics API response.
type SqlStatsResponse struct {
	Items      []SqlStatsItem `json:"items"`
	TotalCount int64          `json:"totalCount"`
}
