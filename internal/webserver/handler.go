package webserver

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"

	"github.com/t-margheim/wordle-pal/internal/pathreview"
)

func NewHandler() (*handler, error) {
	resultTemplate, err := template.ParseFiles("htmx/result.html")
	if err != nil {
		return nil, err
	}

	return &handler{
		svc:            &pathreview.Service{},
		resultTemplate: resultTemplate,
	}, nil
}

type handler struct {
	homePage       []byte
	styles         []byte
	resultTemplate *template.Template
	svc            pathreview.Servicer
}

func (h *handler) Analyze(w http.ResponseWriter, r *http.Request) {
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

	hResp := newAnalyzeResponse(target, resp)
	slog.Debug("analysis prepped", slog.Any("response", hResp))

	err = h.resultTemplate.Execute(w, hResp)
	if err != nil {
		slog.Error("failed to generate response", slog.String("error", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	slog.Info("request map received and parsed", slog.Any("request", reqMap))
}
