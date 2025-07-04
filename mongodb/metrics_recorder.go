package mongodb

type MetricsRecorder interface {
	RecordUpdateSuccess(collection string, count int64)
	RecordUpdateError(collection string, count int64)
	RecordDeleteSuccess(collection string, count int64)
	RecordDeleteError(collection string, count int64)
	RecordProcessLatency(latencyMs int64)
	RecordBulkRequestProcessLatency(latencyMs int64)
}
