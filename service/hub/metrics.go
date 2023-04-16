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
	clientConnGauge      *prometheus.GaugeVec
	apiInsGauge          *prometheus.GaugeVec
	requestCounter       *prometheus.CounterVec
	requestErrCounter    *prometheus.CounterVec
	responseDurHistogram *prometheus.HistogramVec
)

func init() {
	clientConnGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "rohon",
			Subsystem: "risk",
			Name:      "client",
			Help:      "",
		},
		[]string{"count", "total"},
	)
}

// CollectPromStatics collect statics in prometheus
func CollectPromStatics(uri string) error {
	prometheus.MustRegister(clientConnGauge)
	prometheus.MustRegister(apiInsGauge)
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
