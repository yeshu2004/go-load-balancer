package main

import (
    "fmt"
    "net/http"
)

func main() {
    port := ":8004"
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Response from server %s\n", port)
    })
	
    fmt.Printf("Starting mock server on %s\n", port)
    if err := http.ListenAndServe(port, mux); err != nil {
        panic(err)
    }
}