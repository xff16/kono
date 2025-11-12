package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger(buf *bytes.Buffer) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)

	return zap.New(core)
}

func TestLoggerMiddleware_RequestLogged(t *testing.T) {
	buf := new(bytes.Buffer)
	m := &Middleware{
		log:     newTestLogger(buf),
		enabled: true,
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	logOutput := buf.String()

	if !strings.Contains(logOutput, "request started") {
		t.Errorf("expected 'request started' log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "request completed") {
		t.Errorf("expected 'request completed' log, got: %s", logOutput)
	}
}

func TestLoggerMiddleware_Disabled(t *testing.T) {
	buf := new(bytes.Buffer)
	m := &Middleware{
		log:     newTestLogger(buf),
		enabled: false,
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/disabled", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if buf.Len() > 0 {
		t.Errorf("expected no logs, got: %s", buf.String())
	}
}

func TestLoggerMiddleware_LogsBody(t *testing.T) {
	buf := new(bytes.Buffer)
	m := &Middleware{
		log:     newTestLogger(buf),
		logBody: true,
		enabled: true,
	}

	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/body", bytes.NewBufferString(`{"hello":"world"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	logOutput := buf.String()

	if !strings.Contains(logOutput, "request started") ||
		!strings.Contains(logOutput, "request completed") {
		t.Errorf("expected log entries for start and completion, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "body") {
		t.Errorf("expected request body in logs, got: %s", logOutput)
	}
}
