package main

import (
	"time"
)

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func convertToSqlAuditResult(raw *SqlAuditRawResult) SqlAuditResult {
	res := SqlAuditResult{
		CollectTime:           time.Now().UnixMicro(),
		SqlId:                 raw.SqlId,
		FormatSqlId:           raw.FormatSqlId,
		TenantId:              raw.TenantId,
		EffectiveTenantId:     raw.EffectiveTenantId,
		TenantName:            raw.TenantName,
		UserId:                raw.UserId,
		UserName:              raw.UserName,
		DbId:                  raw.DbId,
		DbName:                raw.DbName,
		QuerySql:              raw.QuerySql,
		PlanId:                raw.PlanId,
		TraceId:               raw.TraceId,
		MaxRequestId:          raw.RequestId,
		MinRequestId:          raw.RequestId,
		MaxRequestTime:        raw.RequestTime,
		MinRequestTime:        raw.RequestTime,
		MaxElapsedTime:        raw.ElapsedTime,
		Executions:            1,
		AffectedRows:          raw.AffectedRows,
		ReturnRows:            raw.ReturnRows,
		PartitionCount:        raw.PartitionCount,
		FailCount:             raw.FailCount,
		RetCode4012Count:      raw.RetCode4012Count,
		RetCode4013Count:      raw.RetCode4013Count,
		RetCode5001Count:      raw.RetCode5001Count,
		RetCode5024Count:      raw.RetCode5024Count,
		RetCode5167Count:      raw.RetCode5167Count,
		RetCode5217Count:      raw.RetCode5217Count,
		RetCode6002Count:      raw.RetCode6002Count,
		Event0WaitTime:        raw.Event0WaitTime,
		Event1WaitTime:        raw.Event1WaitTime,
		Event2WaitTime:        raw.Event2WaitTime,
		Event3WaitTime:        raw.Event3WaitTime,
		TotalWaitTimeMicro:    raw.TotalWaitTimeMicro,
		TotalWaits:            raw.TotalWaits,
		RpcCount:              raw.RpcCount,
		PlanTypeLocalCount:    raw.PlanTypeLocalCount,
		PlanTypeRemoteCount:   raw.PlanTypeRemoteCount,
		PlanTypeDistCount:     raw.PlanTypeDistCount,
		InnerSqlCount:         raw.InnerSqlCount,
		ExecutorRpcCount:      raw.ExecutorRpcCount,
		MissPlanCount:         raw.MissPlanCount,
		ElapsedTime:           raw.ElapsedTime,
		NetTime:               raw.NetTime,
		NetWaitTime:           raw.NetWaitTime,
		QueueTime:             raw.QueueTime,
		DecodeTime:            raw.DecodeTime,
		GetPlanTime:           raw.GetPlanTime,
		ExecuteTime:           raw.ExecuteTime,
		ApplicationWaitTime:   raw.ApplicationWaitTime,
		ConcurrencyWaitTime:   raw.ConcurrencyWaitTime,
		UserIoWaitTime:        raw.UserIoWaitTime,
		ScheduleTime:          raw.ScheduleTime,
		RowCacheHit:           raw.RowCacheHit,
		BloomFilterCacheHit:   raw.BloomFilterCacheHit,
		BlockCacheHit:         raw.BlockCacheHit,
		BlockIndexCacheHit:    raw.BlockIndexCacheHit,
		DiskReads:             raw.DiskReads,
		RetryCount:            raw.RetryCount,
		TableScan:             raw.TableScan,
		ConsistencyLevelStrong: raw.ConsistencyLevelStrong,
		ConsistencyLevelWeak:  raw.ConsistencyLevelWeak,
		MemstoreReadRowCount:  raw.MemstoreReadRowCount,
		SSStoreReadRowCount:   raw.SSStoreReadRowCount,
		ExpectedWorkerCount:   raw.ExpectedWorkerCount,
		UsedWorkerCount:       raw.UsedWorkerCount,
		MemoryUsed:            raw.MemoryUsed,
	}
	cpuTime := raw.ExecuteTime - raw.TotalWaitTimeMicro + raw.GetPlanTime
	res.CpuTime = cpuTime
	res.MaxCpuTime = cpuTime
	return res
}

func aggregateGroupBySqlId(one *SqlAuditResult, another *SqlAuditResult) SqlAuditResult {
	return SqlAuditResult{
		CollectTime:           min(one.CollectTime, another.CollectTime),
		SqlId:                 one.SqlId,
		FormatSqlId:           one.FormatSqlId,
		TenantId:              one.TenantId,
		EffectiveTenantId:     one.EffectiveTenantId,
		TenantName:            one.TenantName,
		UserId:                one.UserId,
		UserName:              one.UserName,
		DbId:                  one.DbId,
		DbName:                one.DbName,
		QuerySql:              one.QuerySql, // Keep the first one
		PlanId:                one.PlanId,
		TraceId:               one.TraceId,
		MaxRequestId:          max(one.MaxRequestId, another.MaxRequestId),
		MinRequestId:          min(one.MinRequestId, another.MinRequestId),
		MaxRequestTime:        max(one.MaxRequestTime, another.MaxRequestTime),
		MinRequestTime:        min(one.MinRequestTime, another.MinRequestTime),
		MaxCpuTime:            max(one.MaxCpuTime, another.MaxCpuTime),
		MaxElapsedTime:        max(one.MaxElapsedTime, another.MaxElapsedTime),
		Executions:            one.Executions + another.Executions,
		AffectedRows:          one.AffectedRows + another.AffectedRows,
		ReturnRows:            one.ReturnRows + another.ReturnRows,
		PartitionCount:        one.PartitionCount + another.PartitionCount,
		FailCount:             one.FailCount + another.FailCount,
		RetCode4012Count:      one.RetCode4012Count + another.RetCode4012Count,
		RetCode4013Count:      one.RetCode4013Count + another.RetCode4013Count,
		RetCode5001Count:      one.RetCode5001Count + another.RetCode5001Count,
		RetCode5024Count:      one.RetCode5024Count + another.RetCode5024Count,
		RetCode5167Count:      one.RetCode5167Count + another.RetCode5167Count,
		RetCode5217Count:      one.RetCode5217Count + another.RetCode5217Count,
		RetCode6002Count:      one.RetCode6002Count + another.RetCode6002Count,
		Event0WaitTime:        one.Event0WaitTime + another.Event0WaitTime,
		Event1WaitTime:        one.Event1WaitTime + another.Event1WaitTime,
		Event2WaitTime:        one.Event2WaitTime + another.Event2WaitTime,
		Event3WaitTime:        one.Event3WaitTime + another.Event3WaitTime,
		TotalWaitTimeMicro:    one.TotalWaitTimeMicro + another.TotalWaitTimeMicro,
		TotalWaits:            one.TotalWaits + another.TotalWaits,
		RpcCount:              one.RpcCount + another.RpcCount,
		PlanTypeLocalCount:    one.PlanTypeLocalCount + another.PlanTypeLocalCount,
		PlanTypeRemoteCount:   one.PlanTypeRemoteCount + another.PlanTypeRemoteCount,
		PlanTypeDistCount:     one.PlanTypeDistCount + another.PlanTypeDistCount,
		InnerSqlCount:         one.InnerSqlCount + another.InnerSqlCount,
		ExecutorRpcCount:      one.ExecutorRpcCount + another.ExecutorRpcCount,
		MissPlanCount:         one.MissPlanCount + another.MissPlanCount,
		ElapsedTime:           one.ElapsedTime + another.ElapsedTime,
		NetTime:               one.NetTime + another.NetTime,
		NetWaitTime:           one.NetWaitTime + another.NetWaitTime,
		QueueTime:             one.QueueTime + another.QueueTime,
		DecodeTime:            one.DecodeTime + another.DecodeTime,
		GetPlanTime:           one.GetPlanTime + another.GetPlanTime,
		ExecuteTime:           one.ExecuteTime + another.ExecuteTime,
		CpuTime:               one.CpuTime + another.CpuTime,
		ApplicationWaitTime:   one.ApplicationWaitTime + another.ApplicationWaitTime,
		ConcurrencyWaitTime:   one.ConcurrencyWaitTime + another.ConcurrencyWaitTime,
		UserIoWaitTime:        one.UserIoWaitTime + another.UserIoWaitTime,
		ScheduleTime:          one.ScheduleTime + another.ScheduleTime,
		RowCacheHit:           one.RowCacheHit + another.RowCacheHit,
		BloomFilterCacheHit:   one.BloomFilterCacheHit + another.BloomFilterCacheHit,
		BlockCacheHit:         one.BlockCacheHit + another.BlockCacheHit,
		BlockIndexCacheHit:    one.BlockIndexCacheHit + another.BlockIndexCacheHit,
		DiskReads:             one.DiskReads + another.DiskReads,
		RetryCount:            one.RetryCount + another.RetryCount,
		TableScan:             one.TableScan + another.TableScan,
		ConsistencyLevelStrong: one.ConsistencyLevelStrong + another.ConsistencyLevelStrong,
		ConsistencyLevelWeak:  one.ConsistencyLevelWeak + another.ConsistencyLevelWeak,
		MemstoreReadRowCount:  one.MemstoreReadRowCount + another.MemstoreReadRowCount,
		SSStoreReadRowCount:   one.SSStoreReadRowCount + another.SSStoreReadRowCount,
		ExpectedWorkerCount:   one.ExpectedWorkerCount + another.ExpectedWorkerCount,
		UsedWorkerCount:       one.UsedWorkerCount + another.UsedWorkerCount,
		MemoryUsed:            one.MemoryUsed + another.MemoryUsed,
	}
}
