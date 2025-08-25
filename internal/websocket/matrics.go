// internal/websocket/metrics.go
package websocket

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	metricActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ultaai",
			Subsystem: "ws",
			Name:      "active_connections",
			Help:      "Number of active WebSocket agent connections",
		},
	)

	metricMsgsEnqueued = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ultaai",
			Subsystem: "ws",
			Name:      "messages_enqueued_total",
			Help:      "Total messages successfully enqueued to agent send buffers",
		},
	)

	metricMsgsDropped = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ultaai",
			Subsystem: "ws",
			Name:      "messages_dropped_total",
			Help:      "Total messages dropped due to backpressure or buffer limits",
		},
	)

	metricOfflineBuffered = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ultaai",
			Subsystem: "ws",
			Name:      "offline_buffered_total",
			Help:      "Total messages buffered for offline agents",
		},
	)

	metricOfflineFlushed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ultaai",
			Subsystem: "ws",
			Name:      "offline_flushed_total",
			Help:      "Total messages flushed from offline buffer to agents",
		},
	)
)

func init() {
	// WS metrics
	prometheus.MustRegister(
		metricActiveConnections,
		metricMsgsEnqueued,
		metricMsgsDropped,
		metricOfflineBuffered,
		metricOfflineFlushed,
	)

	// --- New: runtime/process metrics (to prove memory & CPU) ---
	prometheus.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

// Helpers unchanged...
func metricsIncActive()            { metricActiveConnections.Inc() }
func metricsDecActive()            { metricActiveConnections.Dec() }
func metricsEnqueued(n int)        { metricMsgsEnqueued.Add(float64(n)) }
func metricsDropped(n int)         { metricMsgsDropped.Add(float64(n)) }
func metricsOfflineBuffered(n int) { metricOfflineBuffered.Add(float64(n)) }
func metricsOfflineFlushed(n int)  { metricOfflineFlushed.Add(float64(n)) }
