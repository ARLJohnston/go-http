package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"github.com/ARLJohnston/go-http/proto"
	msql "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	client proto.AlbumsClient
	ctx    context.Context
	cfg    msql.Config = msql.Config{
		User:   "root",
		Passwd: "password",
		Net:    "tcp",
		DBName: "album",
	}
)

func TestGrpcCreate(t *testing.T) {
	record := proto.Album{ID: 0, Artist: "Create", Title: "Record", Cover: "Cover", Price: 0}

	id, err := client.Create(ctx, &record)
	if err != nil {
		t.Errorf("Unable to create record: %v", err)
	}
	if id == nil {
		t.Errorf("Create did not return an identifier")
	}

	stream, err := client.Read(ctx, &proto.Nil{})
	if err != nil {
		t.Errorf("Failed to read: %v", err)
	}

	done := make(chan bool)
	defer close(done)
	found := false

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

			if resp.Title == "Record" && resp.Artist == "Create" {
				found = true
				done <- true
				return
			}
		}
	}()

	<-done
	if !found {
		t.Error("Unable to find created record")
	}
}

func TestGrpcRead(t *testing.T) {
	stream, err := client.Read(ctx, &proto.Nil{})
	if err != nil {
		t.Errorf("Failed to read: %v", err)
	}

	done := make(chan bool)
	defer close(done)
	found := false

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

			if resp.Title == "Blue Train" && resp.Artist == "John Coltrane" {
				found = true
				done <- true
				return
			}
		}
	}()

	<-done
	if !found {
		t.Error("Unable to find record")
	}
}

func TestGrpcUpdate(t *testing.T) {
	record := proto.Album{ID: 101, Artist: "Old", Title: "Record", Cover: "Cover", Price: 0}

	id, err := client.Create(ctx, &record)
	if err != nil {
		t.Errorf("Unable to create record: %v", err)
	}

	newRecord := proto.Album{ID: id.Id, Artist: "New", Title: "Record", Cover: "Cover"}
	req := proto.UpdateRequest{OldAlbum: &record, NewAlbum: &newRecord}

	_, err = client.Update(ctx, &req)
	if err != nil {
		t.Errorf("Unable to update record: %v", err)
	}
}

func TestGrpcDelete(t *testing.T) {
	record := proto.Album{ID: 0, Artist: "DeleteMe", Title: "DeleteMe", Cover: "DeleteMe", Price: 0}

	id, err := client.Create(ctx, &record)
	if err != nil {
		t.Errorf("Unable to create record: %v", err)
	}
	if id == nil {
		t.Errorf("Create did not return an identifier")
	}
	record.ID = id.Id

	_, err = client.Delete(ctx, &record)
	if err != nil {
		t.Errorf("Unable to delete record: %v", err)
	}

	stream, err := client.Read(ctx, &proto.Nil{})
	if err != nil {
		t.Errorf("Failed to read: %v", err)
	}

	done := make(chan bool)
	defer close(done)
	found := false

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

			if resp.ID == id.Id {
				found = true
				done <- true
				return
			}
		}
	}()

	<-done
	if found {
		t.Error("Found deleted Record")
	}
}

func TestParseEnvFallback(t *testing.T) {
	got := ParseEnv("veryspecific", "fallback")
	want := "fallback"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}
}

func TestParseEnv(t *testing.T) {
	os.Setenv("env", "val")
	got := ParseEnv("env", "fallback")
	want := "val"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}
}

func TestMain(m *testing.M) {
	ctx = context.Background()
	container, port := startContainer(ctx)

	cfg.Addr = fmt.Sprintf("localhost:%s", port)

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic("Unable to connect to container")
	}

	cli, stopServer := StartServer(ctx)
	client = cli

	ret := m.Run()
	stopServer()
	container.Terminate(ctx)
	os.Exit(ret)
}

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

func StartServer(ctx context.Context) (proto.AlbumsClient, func()) {
	buf := 1024 * 1024
	listener := bufconn.Listen(buf)

	s := grpc.NewServer()
	proto.RegisterAlbumsServer(s, &Server{})
	go func() {
		err := s.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(
		func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))

	client := proto.NewAlbumsClient(conn)

	return client, s.Stop
}
