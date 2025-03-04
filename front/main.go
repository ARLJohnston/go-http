package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/a-h/templ"

	"github.com/ARLJohnston/go-http/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Representation of an Album
type Album struct {
	Id     int64  // SQL Identifier
	Title  string // Album title
	Artist string // Album artist
	Score  int64  // Upvotes
	Cover  string // Link to image of the cover
}

var (
	target string             // Where gRPC client to database is located
	client proto.AlbumsClient // Active gRPC connection to the client

	pageLoads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "front_end_page_loads_total",
		Help: "The total number of times the front end has attempted to be accessed",
	})
	databaseLoads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "front_end_database_loads_total",
		Help: "The total number of database loads from the front end",
	})
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path"})
)

func parseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

// Reads from the database via gRPC and populates the template with streaming
func handleLoad(w http.ResponseWriter, r *http.Request) {
	pageLoads.Inc()
	timer := prometheus.NewTimer(requestDuration.WithLabelValues("/"))
	defer timer.ObserveDuration()

	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		component := unavailable(err.Error())
		templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
		return
	}
	client = proto.NewAlbumsClient(conn)

	stream, err := client.Read(context.Background(), &proto.Nil{})
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
					Id:     resp.Id,
					Title:  resp.Title,
					Artist: resp.Artist,
					Score:  resp.Score,
					Cover:  resp.Cover,
				}
			}
		}()

		<-done
		databaseLoads.Inc()
	}()
	component := grid(data)

	templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
}

// Content of buttons in grid
func post(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.Form.Has("up") {
		id, err := strconv.Atoi(r.Form["up"][0])
		if err != nil {
			log.Printf("unable to decode id %v", err)
			return
		}
		score, err := client.Increment(context.Background(), &proto.Identifier{Id: int64(id)})
		if err != nil {
			log.Printf("cannot receive %v", err)
			return
		}

		component := scoring(id, int(score.Score))
		templ.Handler(component).ServeHTTP(w, r)
	}

	if r.Form.Has("down") {
		id, err := strconv.Atoi(r.Form["down"][0])
		if err != nil {
			log.Printf("unable to decode id %v", err)
			return
		}
		score, err := client.Decrement(context.Background(), &proto.Identifier{Id: int64(id)})
		if err != nil {
			log.Printf("cannot receive %v", err)
			return
		}

		component := scoring(id, int(score.Score))
		templ.Handler(component).ServeHTTP(w, r)
	}
}

// Starts http server with appropriate routes
func main() {
	target = parseEnv("GRPC_TARGET", ":50051")

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", handleLoad)
	http.HandleFunc("/post", post)

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
