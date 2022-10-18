package logging

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type tracingResponseWriter struct {
	next       http.ResponseWriter
	statusCode int
	dumpData   bool
	data       *bytes.Buffer
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
	if w.dumpData {
		w.data.Write(data)
	}
	return w.next.Write(data)
}

func (w *tracingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.next.WriteHeader(statusCode)
}

// HTTPMiddleware installs a logging http.Handler that logs requests and
// selected aspects of responses to 'logger'.
func HTTPMiddleware(logger *zap.Logger, dump bool, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start = time.Now()
			trw   = &tracingResponseWriter{
				dumpData: dump,
				data:     new(bytes.Buffer),
				next:     w,
			}
		)

		if dump {
			// dump request in logs
			reqData, err := io.ReadAll(r.Body)
			if err != nil {
				logger = logger.With(zap.NamedError("req_dump_err", err))
			} else {
				if err := r.Body.Close(); err != nil {
					logger = logger.With(zap.NamedError("req_dump_err", err))
				}
				logger = logger.With(zap.ByteString("req_dump", reqData))

				// replace req.Body with a copy
				r.Body = io.NopCloser(bytes.NewReader(reqData))
			}
		}

		handler.ServeHTTP(trw, r)

		if dump {
			// dump response in logs
			respData, err := io.ReadAll(trw.data)
			if err != nil {
				logger = logger.With(zap.NamedError("resp_dump_err", err))
			} else {
				logger = logger.With(zap.ByteString("resp_dump", respData))
			}
		}

		logger.Info(
			fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.Proto),
			zap.Any("req_headers", r.Header),
			zap.Int("resp_status_code", trw.statusCode),
			zap.String("resp_status_text", http.StatusText(trw.statusCode)),
			zap.String("peer_address", r.RemoteAddr),
			zap.Time("start_time", start),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
