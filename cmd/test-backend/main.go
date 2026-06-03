package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	var port string

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "port=") {
			port = strings.TrimPrefix(arg, "port=")
		}
	}

	addr := net.JoinHostPort("", port)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "backend: %s", addr)
	})

	fmt.Printf("server started on %s\n", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println(err)
	}
}
