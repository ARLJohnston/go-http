package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"

	"github.com/ARLJohnston/go-http/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
	Cover  string
}

var (
	conn   *grpc.ClientConn
	client pb.AlbumsClient

	pageLoads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "front_end_page_loads_total",
		Help: "The total number of times the front end has attempted to be accessed",
	})
	databaseLoads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "front_end_database_loads_total",
		Help: "The total number of database loads from the front end",
	})
)

func parseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func main() {
	target := parseEnv("GRPC_TARGET", ":50051")

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pageLoads.Inc()
		conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			component := unavailable(err.Error())
			templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
			return
		}
		client = pb.NewAlbumsClient(conn)

		stream, err := client.Read(context.Background(), &pb.Nil{})
		if err != nil {
			component := unavailable(err.Error())
			templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
			return
		}
		data := make(chan Album)

		go func() {
			defer close(data)

			done := make(chan bool)
			defer close(done)

			go func() {
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						done <- true
						return
					}
					if err != nil {
						log.Printf("cannot receive %v", err)
						return
					}

					data <- Album{
						ID:     resp.ID,
						Title:  resp.Title,
						Artist: resp.Artist,
						Price:  resp.Price,
						Cover:  resp.Cover,
					}
				}
			}()

			<-done //we will wait until all response is received
			databaseLoads.Inc()
		}()
		component := grid(data)

		templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
	})

	fmt.Println("Listening on : 3000")
	http.ListenAndServe(":3000", nil)
}
