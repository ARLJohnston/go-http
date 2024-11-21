package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ARLJohnston/go-http/pb"
	msql "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
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

func TestCreate(t *testing.T) {
	ctx := context.Background()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	container, port := startContainer(ctx)
	defer container.Terminate(ctx)

	databaseAddress := fmt.Sprintf("localhost:%s", port)

	req := &pb.Album{
		ID:     12,
		Title:  "Hello",
		Artist: "World",
		Price:  5.99,
		Cover:  "",
	}

	cfg := msql.Config{
		User:   "root",
		Passwd: "password",
		Net:    "tcp",
		Addr:   databaseAddress,
		DBName: "album",
	}

	s := server{}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())

	id, err := s.Create(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id.Id == -1 {
		t.Errorf("Did not get an identifier back for creation")
	}
}

// func TestReadContainer(t *testing.T) {
// 	ctx := context.Background()
// 	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
// 	container, port := startContainer(ctx)
// 	defer container.Terminate(ctx)

// 	databaseAddress := fmt.Sprintf("localhost:%s", port)

// 	s := server{
// 		cfg: msql.Config{
// 			User:   "root",
// 			Passwd: "password",
// 			Net:    "tcp",
// 			Addr:   databaseAddress,
// 			DBName: "album",
// 		},
// 	}

// 	_ = s.Read(&pb.Nil{}, nil)
// }

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	container, port := startContainer(ctx)
	defer container.Terminate(ctx)

	databaseAddress := fmt.Sprintf("localhost:%s", port)

	oldAlbum := &pb.Album{
		ID:     1,
		Title:  "Blue Train",
		Artist: "John Coltrane",
		Price:  56.99,
		Cover:  "https://upload.wikimedia.org/wikipedia/en/thumb/6/68/John_Coltrane_-_Blue_Train.jpg/220px-John_Coltrane_-_Blue_Train.jpg",
	}

	newAlbum := &pb.Album{
		ID:     1,
		Title:  "Hello",
		Artist: "World",
		Price:  5.99,
		Cover:  "",
	}

	cfg := msql.Config{
		User:   "root",
		Passwd: "password",
		Net:    "tcp",
		Addr:   databaseAddress,
		DBName: "album",
	}

	s := server{}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())

	update := &pb.UpdateRequest{OldAlbum: oldAlbum, NewAlbum: newAlbum}

	_, err = s.Update(ctx, update)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	//Read
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	container, port := startContainer(ctx)
	defer container.Terminate(ctx)

	databaseAddress := fmt.Sprintf("localhost:%s", port)

	req := &pb.Album{
		ID:     20,
		Title:  "Hello",
		Artist: "World",
		Price:  5.99,
		Cover:  "",
	}

	cfg := msql.Config{
		User:   "root",
		Passwd: "password",
		Net:    "tcp",
		Addr:   databaseAddress,
		DBName: "album",
	}

	s := server{}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())

	id, err := s.Create(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id.Id == -1 {
		t.Errorf("Did not get an identifier back for creation")
	}

	//Check that it is in it

	_, err = s.Delete(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	//Check that it is not in it
}

func TestParseEnv(t *testing.T) {
	os.Setenv("VAR", "variable")
	got := parseEnv("VAR", "fallback")
	want := "variable"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}

}

func TestParseEnvFallback(t *testing.T) {
	got := parseEnv("veryspecific", "fallback")
	want := "fallback"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}

}
