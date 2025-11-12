package main

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressorMiddleware_Gzip(t *testing.T) {
	m := &Middleware{
		enabled: true,
		alg:     "gzip",
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello gzip"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}

	if string(data) != "hello gzip" {
		t.Fatalf("unexpected decompressed body: %s", data)
	}
}

func TestCompressorMiddleware_Deflate(t *testing.T) {
	m := &Middleware{
		enabled: true,
		alg:     "deflate",
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello deflate"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "deflate" {
		t.Fatalf("expected Content-Encoding: deflate, got %s", rec.Header().Get("Content-Encoding"))
	}

	r := flate.NewReader(rec.Body)
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read deflate body: %v", err)
	}

	if string(data) != "hello deflate" {
		t.Fatalf("unexpected decompressed body: %s", data)
	}
}

func TestCompressorMiddleware_NoEncodingHeader(t *testing.T) {
	m := &Middleware{
		enabled: true,
		alg:     "gzip",
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("plain text"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no Content-Encoding, got %s", rec.Header().Get("Content-Encoding"))
	}

	if rec.Body.String() != "plain text" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
