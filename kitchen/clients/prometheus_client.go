package clients

import "github.com/prometheus/client_golang/prometheus"

type PrometheusClient interface {
	RecordCompletedOrders()
	RegisterMetrics()
}
type prometheusClient struct {
}

var totalCompletedOrders = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "completed_orders_total",
		Help: "To get number of Completed Orders",
	},
)

func NewPrometheusClient() PrometheusClient {
	return &prometheusClient{}
}

func (p prometheusClient) RecordCompletedOrders() {
	totalCompletedOrders.Inc()
}

func (p prometheusClient) RegisterMetrics() {
	prometheus.Register(totalCompletedOrders)
}
