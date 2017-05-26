package metrics

import "github.com/prometheus/client_golang/prometheus"

type collector struct {
	metrics []*element
}

func (s *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range s.metrics {
		ch <- c.desc
	}
}

func (s *collector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range s.metrics {
		ch <- c.collect(c.desc)
	}
}

type element struct {
	collect func(desc *prometheus.Desc) prometheus.Metric
	desc    *prometheus.Desc
}
