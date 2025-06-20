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

	s.collectLatencyMetrics(ch, bulkMetric)
	s.collectActionCounters(ch, bulkMetric)
}

func (s *Collector) collectLatencyMetrics(ch chan<- prometheus.Metric, bulkMetric *bulk.Metric) {
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
}

func (s *Collector) collectActionCounters(ch chan<- prometheus.Metric, bulkMetric *bulk.Metric) {
	s.collectCounterMap(ch, bulkMetric.InsertErrorCounter, "insert", "error")
	s.collectCounterMap(ch, bulkMetric.UpdateSuccessCounter, "update", "success")
	s.collectCounterMap(ch, bulkMetric.UpdateErrorCounter, "update", "error")
	s.collectCounterMap(ch, bulkMetric.DeleteSuccessCounter, "delete", "success")
	s.collectCounterMap(ch, bulkMetric.DeleteErrorCounter, "delete", "error")
}

func (s *Collector) collectCounterMap(
	ch chan<- prometheus.Metric, counterMap map[string]int64, actionType, result string,
) {
	for collection, count := range counterMap {
		ch <- prometheus.MustNewConstMetric(
			s.actionCounter,
			prometheus.CounterValue,
			float64(count),
			actionType, result, collection,
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
