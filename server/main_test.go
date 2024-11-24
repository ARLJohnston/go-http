package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"github.com/ARLJohnston/go-http/pb"
	msql "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func startContainer(ctx context.Context) (*mysql.MySQLContainer, string) {
	mysqlC, err := mysql.Run(ctx, "mysql:8.0-bookworm",
		mysql.WithDatabase("album"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts("create-tables.sql"),
	)
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	port, err := mysqlC.MappedPort(ctx, "3306")
	if err != nil {
		log.Fatal(err)
	}

	return mysqlC, port.Port()
}

func StartServer(ctx context.Context) (pb.AlbumsClient, func()) {
	buf := 1024 * 1024
	listener := bufconn.Listen(buf)

	s := grpc.NewServer()
	pb.RegisterAlbumsServer(s, &Server{})
	go func() {
		err := s.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	d := grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		})

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))

	client := pb.NewAlbumsClient(conn)

	return client, s.Stop

}

func TestGrpcRead(t *testing.T) {
	ctx := context.Background()
	container, port := startContainer(ctx)
	defer container.Terminate(ctx)

	databaseAddress := fmt.Sprintf("localhost:%s", port)

	cfg := msql.Config{
		User:   "root",
		Passwd: "password",
		Net:    "tcp",
		Addr:   databaseAddress,
		DBName: "album",
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())

	client, stopServer := StartServer(ctx)
	defer stopServer()

	stream, err := client.Read(ctx, &pb.Nil{})
	if err != nil {
		t.Errorf("Failed to read: %v", err)
	}

	done := make(chan bool)
	defer close(done)
	found := false

	go func() {
		for {
			resp, err := stream.Recv()
			fmt.Println(resp)
			if err == io.EOF {
				done <- true
				return
			}
			if err != nil {
				log.Printf("cannot receive %v", err)
				return
			}

			if resp.Title == "Blue Train" && resp.Artist == "John Coltrane" {
				found = true
			}
		}
	}()

	<-done
	if !found {
		t.Error("Unable to find record")
	}
}

func TestParseEnvFallback(t *testing.T) {
	got := parseEnv("veryspecific", "fallback")
	want := "fallback"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}

}
