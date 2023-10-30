package utils

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// metaSaver decorates http.ResponseWriter to save response meta data - status
// code and response size
type metaSaver struct {
	http.ResponseWriter
	status        int
	statusWritten bool
	bodySize      int
}

// WriteHeader decorates http.ResponseWriter.WriteHeader to save status code
func (s *metaSaver) WriteHeader(statusCode int) {
	if s.statusWritten {
		return
	}

	s.ResponseWriter.WriteHeader(statusCode)

	s.status = statusCode
	s.statusWritten = true
}

// Write decorates http.ResponseWriter.Write to save body size
func (s *metaSaver) Write(data []byte) (int, error) {
	size, err := s.ResponseWriter.Write(data)

	s.bodySize += size

	return size, err //nolint:wrapcheck
}

func HTTPLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(writer http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					logger.Error("error processing a request",
						slog.Any("message", err),
						slog.String("trace", string(debug.Stack())),
					)
				}
			}()

			start := time.Now()

			//nolint:exhaustruct
			wrappedWriter := metaSaver{ResponseWriter: writer}

			next.ServeHTTP(&wrappedWriter, req)

			logger.Info("processed a request",
				slog.Int("status", wrappedWriter.status),
				slog.String("method", req.Method),
				slog.String("path", req.URL.EscapedPath()),
				slog.Int64("duration", time.Since(start).Milliseconds()),
				slog.Int("body_size", wrappedWriter.bodySize),
			)
		}

		return http.HandlerFunc(handler)
	}
}
