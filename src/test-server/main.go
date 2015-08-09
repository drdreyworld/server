package main

import (
	"flag"
	"fmt"
	"net/http"
	"server"
	// "time"
)

func main() {
	srv := server.NewServer()
	srv.Server().Addr = ":8000"

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, "Превееед!")
		w.Header().Set("Connection", "close")
	})

	grace := flag.Bool("grace", false, "")
	flag.Parse()

	srv.Server().Handler = handler
	srv.Start(*grace)
}
