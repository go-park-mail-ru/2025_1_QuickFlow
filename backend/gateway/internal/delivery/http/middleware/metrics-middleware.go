package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"time"

	addr "quickflow/config/micro-addr"
	"quickflow/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rec.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

func (rec *statusRecorder) Flush() {
	if fl, ok := rec.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (rec *statusRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rec.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func getHandlerName(h http.Handler) string {
	var fnPtr uintptr

	switch v := h.(type) {
	case http.HandlerFunc:
		fnPtr = reflect.ValueOf(v).Pointer()
	default:
		fnPtr = reflect.ValueOf(h).Pointer()
	}

	fullName := runtime.FuncForPC(fnPtr).Name()

	lastDot := -1
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '.' {
			lastDot = i
			break
		}
	}
	name := fullName
	if lastDot != -1 && lastDot+1 < len(fullName) {
		name = fullName[lastDot+1:]
	}

	if len(name) > 3 && name[len(name)-3:] == "-fm" {
		name = name[:len(name)-3]
	}

	return name
}

func MetricsMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handlerName := getHandlerName(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rec.statusCode)

			metrics.Hits.WithLabelValues(addr.DefaultGatewayServiceName, handlerName, status).Inc()
			metrics.Timings.WithLabelValues(addr.DefaultGatewayServiceName, handlerName).Observe(duration)

			if rec.statusCode >= 400 {
				metrics.ErrorCounter.WithLabelValues(addr.DefaultGatewayServiceName, handlerName, status).Inc()
			}
		})
	}
}
