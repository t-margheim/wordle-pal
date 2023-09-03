package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/t-margheim/wordle-pal/internal/webserver"
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

	handler, err := webserver.NewHandler()
	if err != nil {
		log.Fatalln(err)
	}

	fs := http.FileServer(http.Dir("htmx/static"))
	http.Handle("/static/", addLogMiddleware(http.StripPrefix("/static/", fs)))
	http.Handle("/analyze", addLogMiddleware(http.HandlerFunc(handler.Analyze)))
	http.Handle("/", addLogMiddleware(fs))

	err = http.ListenAndServe(*listen, nil)
	log.Fatalln(err)
}
