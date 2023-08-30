cli:
    go run ./cmd/cli -t EXACT -p arose,cheat,enact,exact 

cli-debug:
    go run ./cmd/cli -t EXACT -p arose,cheat,enact,exact -d

htmx:
    go run ./cmd/server --listen :8080 --debug
