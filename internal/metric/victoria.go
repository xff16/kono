package metric

import (
	"strconv"
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
	ResponsesTotal      map[string]*metrics.Counter
	RequestsInFlight    *metrics.Gauge
	FailedRequestsTotal map[FailReason]*metrics.Counter
}

func New() *Metrics {
	return &Metrics{
		RequestsTotal:    metrics.GetOrCreateCounter("tokka_requests_total"),
		RequestsDuration: metrics.GetOrCreateSummary("tokka_requests_duration_seconds"),
		ResponsesTotal: map[string]*metrics.Counter{
			"200":   metrics.GetOrCreateCounter(`tokka_responses_total{status="200"}`),
			"301":   metrics.GetOrCreateCounter(`tokka_responses_total{status="301"}`),
			"401":   metrics.GetOrCreateCounter(`tokka_responses_total{status="401"}`),
			"403":   metrics.GetOrCreateCounter(`tokka_responses_total{status="403"}`),
			"404":   metrics.GetOrCreateCounter(`tokka_responses_total{status="404"}`),
			"500":   metrics.GetOrCreateCounter(`tokka_responses_total{status="500"}`),
			"502":   metrics.GetOrCreateCounter(`tokka_responses_total{status="502"}`),
			"other": metrics.GetOrCreateCounter(`tokka_responses_total{status="other"}`),
		},
		RequestsInFlight: metrics.GetOrCreateGauge("tokka_requests_in_flight", nil),
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

func (m *Metrics) UpdateRequestsDuration(start time.Time) {
	m.RequestsDuration.UpdateDuration(start)
}

func (m *Metrics) IncResponsesTotal(status int) {
	if _, ok := m.ResponsesTotal[strconv.Itoa(status)]; !ok {
		m.ResponsesTotal["other"].Inc()
		return
	}

	m.ResponsesTotal[strconv.Itoa(status)].Inc()
}

func (m *Metrics) IncRequestsInFlight() {
	m.RequestsInFlight.Inc()
}

func (m *Metrics) DecRequestsInFlight() {
	m.RequestsInFlight.Dec()
}

func (m *Metrics) IncFailedRequestsTotal(reason FailReason) {
	if _, ok := m.FailedRequestsTotal[reason]; !ok {
		return
	}

	m.FailedRequestsTotal[reason].Inc()
}
