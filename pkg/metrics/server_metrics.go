package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	serverNamespace = "gearmanserver"
)

var inc = 100.0

func NewServerCollector() prometheus.Collector {
	return &collector{
		metrics: []*element{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(serverNamespace, "", "hello_world"),
					"Help Text Here",
					[]string{"hello", "world"}, map[string]string{"ok": "this"},
				),
				collect: func(d *prometheus.Desc) prometheus.Metric {
					inc++
					return prometheus.MustNewConstMetric(d, prometheus.GaugeValue, inc, "hello", "world")
				},
			},
		},
	}
}
