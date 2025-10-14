package kairyu

import (
	"io"
	"net/http"
	"sync"
)

type dispatcher interface {
	dispatch(route *Route, original *http.Request) [][]byte
}

type defaultDispatcher struct{}

func (d *defaultDispatcher) dispatch(route *Route, original *http.Request) [][]byte {
	var wg sync.WaitGroup

	results := make([][]byte, len(route.Backends))
	client := &http.Client{}

	for i, b := range route.Backends {
		wg.Add(1)
		go func(i int, b Backend) {
			defer wg.Done()

			m := b.Method
			if m == "" {
				// Fallback method.
				m = original.Method
			}

			req, err := http.NewRequest(m, b.URL, nil)
			if err != nil {
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				return
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			results[i] = body
		}(i, b)
	}

	wg.Wait()

	return results
}
