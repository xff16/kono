package tokka

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"

	"go.uber.org/zap"

	"github.com/starwalkn/tokka/internal/metric"
)

const maxBodySize = 5 << 20 // 5MB

type dispatcher interface {
	dispatch(route *Route, original *http.Request) []UpstreamResponse
}

type defaultDispatcher struct {
	log     *zap.Logger
	metrics metric.Metrics
}

// dispatch sends the incoming request to all upstreams configured for the given route
// and collects their responses. It applies the upstream policies and metrics, and returns
// a slice of UpstreamResponse.
//
// The dispatching pipeline is as follows:
//
// 1. Reads and limits the request body to `maxBodySize` (default 5MB).
//    - If the body exceeds the limit, returns nil to indicate failure.
// 2. Launches a goroutine for each upstream, sending the request concurrently.
// 3. Each upstream response is validated against its policy:
//    - AllowedStatuses: the response status must be in the allowed list.
//    - RequireBody: the response must have a non-empty body if required.
//    - MapStatusCodes: optionally remaps status codes.
// 4. Any policy violations are converted to UpstreamError and metrics are updated.
// 5. All upstream responses are written to a preallocated slice at their respective index.
// 6. The dispatcher waits for all goroutines to complete before returning the results.
//
// Notes / considerations:
//
// - This dispatcher reads the full request body into memory for each upstream.
//   Large bodies or many upstreams may increase memory usage significantly.
// - Returns nil for body read errors or size violations, which must be checked by the caller.
// - Goroutines write directly to a shared slice; current implementation is safe because
//   each index is unique, but panics in goroutines may affect stability.
// - No global timeout is applied; dispatcher relies on the request context for cancellation.
// - Errors are aggregated using errors.Join, which may become large for multiple violations.
func (d *defaultDispatcher) dispatch(route *Route, original *http.Request) []UpstreamResponse {
	results := make([]UpstreamResponse, len(route.Upstreams))

	originalBody, readErr := io.ReadAll(io.LimitReader(original.Body, maxBodySize+1))
	if readErr != nil {
		d.log.Error("cannot read body", zap.Error(readErr))
		return nil
	}
	if readErr = original.Body.Close(); readErr != nil {
		d.log.Warn("cannot close original request body", zap.Error(readErr))
	}

	if len(originalBody) > maxBodySize {
		d.metrics.IncFailedRequestsTotal(metric.FailReasonBodyTooLarge)
		return nil
	}

	var wg sync.WaitGroup

	for i, u := range route.Upstreams {
		wg.Add(1)

		go func(i int, u Upstream, originalBody []byte) {
			defer wg.Done()

			upstreamPolicy := u.Policy()

			resp := u.Call(original.Context(), original, originalBody, upstreamPolicy.RetryPolicy)
			if resp.Err != nil {
				d.metrics.IncFailedRequestsTotal(metric.FailReasonUpstreamError)
				d.log.Error("upstream request failed",
					zap.String("name", u.Name()),
					zap.Error(resp.Err.Unwrap()),
				)
			}

			if resp.Status != 0 {
				d.metrics.IncResponsesTotal(resp.Status)
			}

			var errs []error

			if upstreamPolicy.RequireBody && len(resp.Body) == 0 {
				errs = append(errs, errors.New("empty body not allowed by upstream policy"))
			}

			if mapped, ok := upstreamPolicy.MapStatusCodes[resp.Status]; ok {
				resp.Status = mapped
			}

			if len(upstreamPolicy.AllowedStatuses) > 0 && !slices.Contains(upstreamPolicy.AllowedStatuses, resp.Status) {
				errs = append(errs, fmt.Errorf("status %d not allowed by upstream policy", resp.Status))
			}

			if len(errs) > 0 {
				d.metrics.IncFailedRequestsTotal(metric.FailReasonPolicyViolation)

				if resp.Err == nil {
					resp.Err = &UpstreamError{
						Err: errors.Join(errs...),
					}
				} else {
					resp.Err.Err = errors.Join(resp.Err.Err, errors.Join(errs...))
				}
			}

			results[i] = *resp
		}(i, u, originalBody)
	}

	wg.Wait()

	return results
}
