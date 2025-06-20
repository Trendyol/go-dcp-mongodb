package metric

import (
	"github.com/Trendyol/go-dcp-mongodb/mongodb/bulk"
	"github.com/Trendyol/go-dcp/helpers"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	bulk *bulk.Bulk

	processLatency            *prometheus.Desc
	bulkRequestProcessLatency *prometheus.Desc
	actionCounter             *prometheus.Desc
}

func (s *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(s, ch)
}

func (s *Collector) Collect(ch chan<- prometheus.Metric) {
	s.bulk.LockMetrics()
	defer s.bulk.UnlockMetrics()

	bulkMetric := s.bulk.GetMetric()

	ch <- prometheus.MustNewConstMetric(
		s.processLatency,
		prometheus.GaugeValue,
		float64(bulkMetric.ProcessLatencyMs),
		[]string{}...,
	)

	ch <- prometheus.MustNewConstMetric(
		s.bulkRequestProcessLatency,
		prometheus.GaugeValue,
		float64(bulkMetric.BulkRequestProcessLatencyMs),
		[]string{}...,
	)

	for collection, count := range bulkMetric.InsertErrorCounter {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			"insert", "error", collection,
		)
	}

	for collection, count := range bulkMetric.UpdateSuccessCounter {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			"update", "success", collection,
		)
	}

	for collection, count := range bulkMetric.UpdateErrorCounter {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			"update", "error", collection,
		)
	}

	for collection, count := range bulkMetric.DeleteSuccessCounter {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			"delete", "success", collection,
		)
	}

	for collection, count := range bulkMetric.DeleteErrorCounter {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			"delete", "error", collection,
		)
	}
}

func NewMetricCollector(bulk *bulk.Bulk) *Collector {
	return &Collector{
		bulk: bulk,

		processLatency: prometheus.NewDesc(
			prometheus.BuildFQName(helpers.Name, "mongodb_connector_latency_ms", "current"),
			"Mongodb connector latency ms",
			[]string{},
			nil,
		),

		bulkRequestProcessLatency: prometheus.NewDesc(
			prometheus.BuildFQName(helpers.Name, "mongodb_connector_bulk_request_process_latency_ms", "current"),
			"Mongodb connector bulk request process latency ms",
			[]string{},
			nil,
		),

		actionCounter: prometheus.NewDesc(
			prometheus.BuildFQName(helpers.Name, "mongodb_connector_action_total", "current"),
			"Mongodb connector action counter",
			[]string{"action_type", "result", "database_name"},
			nil,
		),
	}
}
