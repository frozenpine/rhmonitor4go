package hub

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultListen = "0.0.0.0"
	defaultPort   = 9811
)

var (
	promListen string
	promPort   int
)

type promMetrics struct {
	clientConnGauge      prometheus.GaugeVec
	apiInsGauge          prometheus.GaugeVec
	requestCounter       prometheus.CounterVec
	requestErrCounter    prometheus.CounterVec
	responseDurHistogram prometheus.HistogramVec
}
