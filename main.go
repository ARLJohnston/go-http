package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func get(w http.ResponseWriter, req *http.Request) {
	f, err := os.ReadFile(req.PathValue("file"))
	if err != nil {
		fmt.Fprint(w, "GET: An err occurred: ", err)
	} else {
		fmt.Fprint(w, string(f))
	}
}

func put(w http.ResponseWriter, req *http.Request) {
	f := req.PathValue("file")

	err := os.WriteFile(f, []byte("Example text"), 0666)
	if err != nil {
		fmt.Fprint(w, "PUT: An err occurred: ", err)
	} else {
		fmt.Fprint(w, "File was created")
	}
}

func greet(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello World")
}

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main() {
	err := os.Mkdir("data", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	recordMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// router.HandleFunc("GET /{file}", get)
	// router.HandleFunc("PUT /{file}", put)
	http.HandleFunc("/hello", greet)

	http.ListenAndServe(":8080", nil)
}
