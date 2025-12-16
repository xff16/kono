package metric

import (
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type FailReason string

const (
	FailReasonGatewayError   FailReason = "gateway_error"
	FailReasonUpstreamError  FailReason = "upstream_error"
	FailReasonNoMatchedRoute FailReason = "no_matched_route"
)

type Metrics struct {
	RequestsTotal       *metrics.Counter
	RequestsDuration    *metrics.Summary
	FailedRequestsTotal map[FailReason]*metrics.Counter
}

func New() *Metrics {
	return &Metrics{
		RequestsTotal:    metrics.GetOrCreateCounter("tokka_requests_total"),
		RequestsDuration: metrics.GetOrCreateSummary("tokka_requests_duration_seconds"),
		FailedRequestsTotal: map[FailReason]*metrics.Counter{
			FailReasonGatewayError:   metrics.GetOrCreateCounter(`tokka_failed_requests_total{reason="gateway_error"}`),
			FailReasonUpstreamError:  metrics.GetOrCreateCounter(`tokka_failed_requests_total{reason="upstream_error"}`),
			FailReasonNoMatchedRoute: metrics.GetOrCreateCounter(`tokka_failed_requests_total{reason="no_matched_route"}`),
		},
	}
}

func (m *Metrics) IncRequestsTotal() {
	m.RequestsTotal.Inc()
}

func (m *Metrics) IncFailedRequestsTotal(reason FailReason) {
	if _, ok := m.FailedRequestsTotal[reason]; !ok {
		return
	}

	m.FailedRequestsTotal[reason].Inc()
}

func (m *Metrics) UpdateRequestsDuration(start time.Time) {
	m.RequestsDuration.UpdateDuration(start)
}
