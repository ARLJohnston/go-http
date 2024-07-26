package main

import (
	"fmt"
	"net/http"
)

const (
	greeting = "Hello, "
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, greeting+"World!")
}

func main() {
	http.HandleFunc("/hello", hello)

	http.ListenAndServe(":8080", nil)
}
