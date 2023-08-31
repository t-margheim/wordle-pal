package main

import (
	"log/slog"
	"net/http"
	"os"
)

func initLogger() {
	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}),
	))
}

type loggingResponseWriter struct {
	internalWriter http.ResponseWriter
	statusCode     int
	byteCount      int
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.internalWriter.Header()
}

func (lrw *loggingResponseWriter) Write(bb []byte) (int, error) {
	byteCount, err := lrw.internalWriter.Write(bb)
	lrw.byteCount = byteCount
	return byteCount, err
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.internalWriter.WriteHeader(code)
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		internalWriter: w,
		statusCode:     http.StatusOK,
	}
}

func addLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("serving request",
			slog.String("request_url", r.URL.String()),
		)
		slogWriter := newLoggingResponseWriter(w)
		next.ServeHTTP(slogWriter, r)
		slog.Debug("sending response",
			slog.Int("status", slogWriter.statusCode),
			slog.Int("size", slogWriter.byteCount),
		)
	})
}
