package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	Collectors() []prometheus.Collector
	RefreshMetrics() error
}
