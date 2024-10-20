package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ARLJohnston/go-http/pb"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func TestStartContainer(t *testing.T) {
	ctx := context.Background()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	container, port := startContainer(ctx)
	defer container.Terminate(ctx)

	// cfg = mysql.Config{
	// 	User:   "root",
	// 	Passwd: "password",
	// 	Net:    "tcp",
	// 	Addr:   "localhost:3306",
	// 	DBName: "album",
	// }
	databaseAddress := fmt.Sprintf("localhost:%s", port)
	os.Setenv("MYSQL_DATABASE_ADDRESS", databaseAddress)

	// s := server{}
	// s.cfg = msql.Config{
	// 	User:   "root",
	// 	Passwd: "password",
	// 	Net:    "tcp",
	// 	Addr:   databaseAddress,
	// 	DBName: "album",
	// }
	go main()

	req := &pb.Album{
		ID:     1,
		Title:  "Hello",
		Artist: "World",
		Price:  5.99,
		Cover:  "",
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf(err.Error())
	}
	client := pb.NewAlbumsClient(conn)
	id, err := client.Create(ctx, req)

	// resp, err := s.Create(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id.Id == -1 {
		t.Errorf("Did not get an identifier back for creation")
	}

}
