package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
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
