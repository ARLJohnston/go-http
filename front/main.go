package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/a-h/templ"

	"github.com/ARLJohnston/go-http/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
)

func parseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func main() {
	//target := parseEnv("GRPC_TARGET", ":8080")
	target := "client:50051"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			msg := err.Error()
			component := unavailable(msg)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		client = pb.NewAlbumsClient(conn)

		stream, err := client.Read(context.Background(), &pb.Nil{})
		if err != nil {
			msg := err.Error()
			component := unavailable(msg)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		// Create a channel to send data to the template.
		data := make(chan Album)
		//Prevent infinite loading
		done := make(chan bool)

		go func() {
			defer close(data)
			for {
				select {
				case <-r.Context().Done():
					return
				default:
					resp, err := stream.Recv()
					if err == io.EOF {
						done <- true
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
			}
		}()

		// Pass the channel to the template.
		component := grid(data)

		// Serve using the streaming mode of the handler.
		templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
		<-done
	})

	fmt.Println("Listening on : 3000")
	http.ListenAndServe(":3000", nil)
}
