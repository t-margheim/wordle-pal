package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/t-margheim/wordle-pal/pkg/pathreview"
)

func NewHandler() (http.Handler, error) {
	indexFile, err := os.Open("htmx/index.html")
	if err != nil {
		return nil, err
	}

	homePage, err := io.ReadAll(indexFile)
	if err != nil {
		return nil, err
	}

	resultTemplate, err := template.ParseFiles("htmx/result.html")
	if err != nil {
		return nil, err
	}

	slog.Info("loaded html home page", slog.Int("size", len(homePage)))

	return &handler{
		homePage:       homePage,
		svc:            &pathreview.Service{},
		resultTemplate: resultTemplate,
	}, nil
}

type handler struct {
	homePage       []byte
	resultTemplate *template.Template
	svc            pathreview.Servicer
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

		target := reqMap.Get("answer")
		if target == "" {
			slog.Error("invalid request message", slog.String("error", "request must have answer value"))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var path []string
		keyBase := "guess_%d"
		for i := 0; i < 6; i++ {
			pathElement := reqMap.Get(fmt.Sprintf(keyBase, i+1))
			if len(pathElement) == 5 {
				path = append(path, pathElement)
			}
		}

		resp, err := h.svc.ReviewPath(pathreview.PathRequest{
			Target: target,
			Path:   path,
		})
		if err != nil {
			slog.Error("failed to review path", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		type htmlResponse struct {
			Response pathreview.PathResponse
			Target   string
		}
		hResp := htmlResponse{
			Response: resp,
			Target:   target,
		}
		err = h.resultTemplate.Execute(w, hResp)
		if err != nil {
			slog.Error("failed to generate response", slog.String("error", err.Error()))
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
