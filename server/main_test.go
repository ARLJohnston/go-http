package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/ARLJohnston/go-http/pb"
	"github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startContainer(ctx context.Context) (testcontainers.Container, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "albums",
		},
		Mounts:     make(testcontainers.ContainerMounts, 0),
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL").WithStartupTimeout(60 * time.Second),
	}

	mysqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	cleanup := func() {
		err := mysqlC.Terminate(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return mysqlC, cleanup
}

func TestStartContainer(t *testing.T) {
	ctx := context.Background()
	_, cleanup := startContainer(ctx)
	defer cleanup()

	cfg = mysql.Config{
		User:   parseEnv("MYSQL_USER", "root"),
		Passwd: parseEnv("MYSQL_USER_PASSWORD", "password"),
		Net:    parseEnv("MYSQL_NETWORK_PROTOCOL", "tcp"),
		Addr:   parseEnv("MYSQL_DATABASE_ADDRESS", "localhost:3306"),
		DBName: parseEnv("MYSQL_DATABASE_NAME", "album"),
	}

	s := server{}

	req := &pb.Album{
		ID:     1,
		Title:  "Hello",
		Artist: "World",
		Price:  5.99,
		Cover:  "",
	}

	resp, err := s.Create(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if resp.Id == -1 {
		t.Errorf("Did not get an identifier back for creation")
	}

}
