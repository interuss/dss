package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type tracingResponseWriter struct {
	next       http.ResponseWriter
	statusCode int
}

func (w *tracingResponseWriter) Header() http.Header {
	return w.next.Header()
}

func (w *tracingResponseWriter) Write(data []byte) (int, error) {
	// While this looks slightly ugly, it follows the specification
	// for an http.ResponseWriter according to:
	//   https://golang.org/pkg/net/http/#ResponseWriter
	if w.statusCode == 0 {
		w.statusCode = 200
	}
	return w.next.Write(data)
}

func (w *tracingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.next.WriteHeader(statusCode)
}

// HTTPMiddleware installs a logging http.Handler that logs requests and
// selected aspects of responses to 'logger'.
func HTTPMiddleware(logger *zap.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start = time.Now()
			trw   = &tracingResponseWriter{
				next: w,
			}
		)

		handler.ServeHTTP(trw, r)

		logger.Info(
			fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.Proto),
			zap.Any("req_headers", r.Header),
			zap.Int("resp_status_code", trw.statusCode),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
