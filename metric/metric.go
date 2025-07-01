package metric

import (
	"github.com/Trendyol/go-dcp-mongodb/mongodb"
	"github.com/Trendyol/go-dcp/helpers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	updateCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(helpers.Name, "mongodb_connector_update_operations", "total"),
			Help: "The total number of update operations",
		},
		[]string{"collection", "status"}, // status: success, error
	)

	deleteCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(helpers.Name, "mongodb_connector_delete_operations", "total"),
			Help: "The total number of delete operations",
		},
		[]string{"collection", "status"}, // status: success, error
	)

	processLatencyGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(helpers.Name, "mongodb_connector_latency_ms", "current"),
			Help: "Process latency in milliseconds",
		},
	)

	bulkRequestProcessLatencyGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(helpers.Name, "mongodb_connector_bulk_request_process_latency_ms", "current"),
			Help: "Bulk request process latency in milliseconds",
		},
	)
)

type PrometheusMetricsRecorder struct{}

func NewMetricsRecorder() mongodb.MetricsRecorder {
	return &PrometheusMetricsRecorder{}
}

func (m *PrometheusMetricsRecorder) RecordUpdateSuccess(collection string, count int64) {
	updateCounter.WithLabelValues(collection, "success").Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordUpdateError(collection string, count int64) {
	updateCounter.WithLabelValues(collection, "error").Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordDeleteSuccess(collection string, count int64) {
	deleteCounter.WithLabelValues(collection, "success").Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordDeleteError(collection string, count int64) {
	deleteCounter.WithLabelValues(collection, "error").Add(float64(count))
}

func (m *PrometheusMetricsRecorder) RecordProcessLatency(latencyMs int64) {
	processLatencyGauge.Set(float64(latencyMs))
}

func (m *PrometheusMetricsRecorder) RecordBulkRequestProcessLatency(latencyMs int64) {
	bulkRequestProcessLatencyGauge.Set(float64(latencyMs))
}
