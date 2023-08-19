package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/t-margheim/wordle-pal/pkg/pathreview"
)

func main() {
	initLogger()

	slog.Debug("command started",
		slog.Bool("debug", debug),
		slog.String("target", target),
		slog.Any("path", path),
	)

	svc := &pathreview.Service{}

	res, err := svc.ReviewPath(pathreview.PathRequest{
		Target: target,
		Path:   path,
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}

	slog.Debug("results returned", slog.Int("count", len(res.GuessResults)))
	for _, r := range res.GuessResults {
		delta := r.PreviousWordCount - r.NewWordCount
		deltaPercent := float64(delta) / float64(r.PreviousWordCount) * 100
		fmt.Printf("Guess %q removed %d words (%02f%%) from the remaining pool.\n", r.Guess, delta, deltaPercent)
	}
}
