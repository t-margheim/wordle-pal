package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
	debug  = flag.Bool("debug", false, "include debug logs in output")
)

func main() {
	flag.Parse()
	initLogger()
	slog.Info("server started",
		slog.String("listen_address", *listen),
	)

	handler, err := NewHandler()
	if err != nil {
		log.Fatalln(err)
	}

	err = http.ListenAndServe(*listen, addLogMiddleware(handler))
	log.Fatalln(err)
}

func NewHandler() (http.Handler, error) {
	indexFile, err := os.Open("htmx/index.html")
	if err != nil {
		return nil, err
	}

	homePage, err := io.ReadAll(indexFile)
	if err != nil {
		return nil, err
	}

	slog.Info("loaded html home page", slog.Int("size", len(homePage)))

	return &handler{
		homePage: homePage,
	}, nil
}

type handler struct {
	homePage []byte
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r)
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	pathMatch := "no match"
	var err error
	switch r.URL.Path {
	case "", "/", "/index.html":
		pathMatch = "home"
		_, err = w.Write(h.homePage)

	case "/start":
		pathMatch = "start"
		f, err := os.Open("./htmx/solve.html")
		if err != nil {
			slog.Error("failed to open answer file", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			slog.Error("failed to read answer file", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bb)

	case "/analyze":
		reqBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("failed to read request body", slog.String("error", err.Error()))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		slog.Debug("analyze request received",
			slog.String("raw_data", string(reqBytes)),
		)
		reqMap, err := url.ParseQuery(string(reqBytes))
		if err != nil {
			slog.Error("failed to parse request body", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		slog.Info("request map received and parsed", slog.Any("request", reqMap))

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
	slog.Debug("request path match",
		slog.String("path", r.URL.Path),
		slog.String("internal", pathMatch),
	)
	if err != nil {
		slog.Error("failed to write response", slog.String("err", err.Error()))
	}
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
