package model

type SqlDetailRequest struct {
    StartTime      int64    `json:"startTime,omitempty"`
    EndTime        int64    `json:"endTime,omitempty"`
    SqlId          string   `json:"sqlId" binding:"required"`
    Interval       int      `json:"interval" binding:"required"`
    LatencyColumns []string `json:"latencyColumns"`
}

type PlanTypeTrend struct {
    Time        int64   `json:"time"`
    Local       float64 `json:"local"`
    Remote      float64 `json:"remote"`
    Distributed float64 `json:"distributed"`
}

type LatencyTrendItem struct {
    Time  int64            `json:"time"`
    Value map[string]float64 `json:"value"`
}

type SqlDetailResponse struct {
    ExecutionTrend []PlanTypeTrend    `json:"executionTrend"`
    LatencyTrend   []LatencyTrendItem `json:"latencyTrend"`
}
