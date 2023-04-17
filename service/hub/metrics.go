package hub

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultPromStaticsURI = "/metrics"
)

var (
	defines = [2]string{"rohon", "risk"}

	streamConnGauge      *prometheus.GaugeVec
	requestCounter       *prometheus.CounterVec
	requestErrCounter    *prometheus.CounterVec
	responseDurHistogram *prometheus.HistogramVec
)

func init() {
	streamConnGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: defines[0],
			Subsystem: defines[1],
			Name:      "stream",
			Help:      "Risk stream client total number",
		},
		[]string{"method"},
	)

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: defines[0],
			Subsystem: defines[1],
			Name:      "request",
			Help:      "Risk api request count",
		},
		[]string{"method"},
	)

	requestErrCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: defines[0],
			Subsystem: defines[1],
			Name:      "reqErr",
			Help:      "Risk api request error count",
		},
		[]string{"method"},
	)

	responseDurHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: defines[0],
			Subsystem: defines[1],
			Name:      "reqDur",
			Help:      "Risk api request duration histogram",
			Buckets:   []float64{100, 200, 300, 500, 750, 1000, 1500, 2000, 2500, 5000, 10000},
		},
		[]string{"method"},
	)
}

// CollectPromStatics collect statics in prometheus
func CollectPromStatics(uri string) error {
	prometheus.MustRegister(streamConnGauge)
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestErrCounter)
	prometheus.MustRegister(responseDurHistogram)

	if uri == "" {
		uri = DefaultPromStaticsURI
	}

	if err := registerHandler(uri, promhttp.Handler()); err != nil {
		return err
	}

	log.Printf("Prometheus URI registered, expose metrics at \"%s\"", uri)

	return nil
}
