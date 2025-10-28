package bravka

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	internalError = "Internal Error"
)

type dispatcher interface {
	dispatch(route *Route, original *http.Request) [][]byte
}

type defaultDispatcher struct {
	log *zap.Logger
}

func (d *defaultDispatcher) dispatch(route *Route, original *http.Request) [][]byte {
	var wg sync.WaitGroup

	results := make([][]byte, len(route.Backends))
	client := &http.Client{}

	originalBody, err := io.ReadAll(original.Body)
	if err != nil {
		d.log.Error("cannot read body", zap.Error(err))
		return nil
	}
	if err = original.Body.Close(); err != nil {
		d.log.Warn("cannot close original request body", zap.Error(err))
	}

	for i, b := range route.Backends {
		wg.Add(1)

		go func(i int, b Backend, originalBody []byte) {
			defer wg.Done()

			m := b.Method
			if m == "" {
				// Fallback method.
				m = original.Method
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(b.Timeout))
			defer cancel()

			// Send request body only for body-acceptable methods requests.
			if m != http.MethodPost && m != http.MethodPut && m != http.MethodPatch {
				originalBody = nil
			}

			req, err := http.NewRequestWithContext(ctx, m, b.URL, bytes.NewReader(originalBody))
			if err != nil {
				results[i] = []byte(internalError)
				return
			}

			q := req.URL.Query()
			for _, fqs := range b.ForwardQueryStrings {
				if fqs == "*" {
					q = original.URL.Query()
					break
				}

				if original.URL.Query().Get(fqs) == "" {
					continue
				}

				q.Add(fqs, original.URL.Query().Get(fqs))
			}
			req.URL.RawQuery = q.Encode()

			// TODO:: implement headers pattern aka "X-*" (forward all headers which starts with X-).
			// Set forwarding headers.
			for _, fw := range b.ForwardHeaders {
				if fw == "*" {
					req.Header = original.Header.Clone()
					break
				}

				if original.Header.Get(fw) == "" {
					continue
				}

				req.Header.Add(http.CanonicalHeaderKey(fw), original.Header.Get(fw))
			}

			// Rewrite headers which exists in backend headers configuration (rewriting only forwarded headers).
			for header, value := range b.Headers {
				if !slices.Contains(b.ForwardHeaders, header) {
					continue
				}

				req.Header.Set(http.CanonicalHeaderKey(header), value)
			}

			req.Header.Set("Content-Type", original.Header.Get("Content-Type"))

			d.log.Info("dispatching request", zap.String("method", m), zap.String("url", req.URL.String()))

			resp, err := client.Do(req)
			if err != nil {
				results[i] = []byte(internalError)

				d.log.Error("backend request failed", zap.String("method", m), zap.Error(err))
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				results[i] = []byte(internalError)

				d.log.Error("cannot read backend response body", zap.Error(err))
				return
			}

			if err = resp.Body.Close(); err != nil {
				d.log.Warn("cannot close backend response body", zap.Error(err))
			}

			results[i] = body
		}(i, b, originalBody)
	}

	wg.Wait()

	return results
}
