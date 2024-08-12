package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func get(w http.ResponseWriter, req *http.Request) {
	f, err := os.ReadFile("." + req.URL.Path)
	if err != nil {
		fmt.Fprint(w, "An err occurred: ", err)
	} else {
		fmt.Fprint(w, string(f))
	}
}

func main() {
	os.Mkdir("data", os.ModePerm)

	err := os.WriteFile("data/ex", []byte("Example text"), 0666)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /", get)

	http.ListenAndServe(":8080", nil)
}
