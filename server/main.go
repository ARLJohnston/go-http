// Package main implements a gRPC server for interaction with a MYSQL database
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"database/sql"

	"github.com/go-sql-driver/mysql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ARLJohnston/go-http/proto"
)

var (
	db *sql.DB // Handle to the database

	opsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_client_started_ops_total",
		Help: "The total number of database calls by the gRPC client",
	})
	opsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_client_failed_ops_total",
		Help: "The total number of failed database calls by the gRPC client",
	})
	opsSucceeded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_client_successful_ops_total",
		Help: "The total number of successful database calls by the gRPC client",
	})
)

// Returns value of environment variable if it is set, otherwise returns fallback
func ParseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

type Server struct {
	proto.UnimplementedAlbumsServer
}

// Creates an Album alb in the db, returns the SQL identifier for that album in the database
func (s *Server) Create(ctx context.Context, alb *proto.Album) (*proto.Identifier, error) {
	opsStarted.Inc()

	result, err := db.Exec("INSERT INTO album (title, artist, score, cover) VALUES (?, ?, ?, ?)", alb.Title, alb.Artist, alb.Score, alb.Cover)
	if err != nil {
		opsFailed.Inc()
		log.Println("Create failed:" + err.Error())
		return nil, status.Error(
			codes.Internal, "Create failed: "+err.Error(),
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		opsFailed.Inc()
		log.Println("Failed to get last insert id: " + err.Error())
		return nil, status.Error(
			codes.NotFound, "Failed to get last insert id: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Identifier{Id: id}, nil
}

// Opens stream for a streaming read of every album in the database
func (s *Server) Read(_ *proto.Nil, stream proto.Albums_ReadServer) error {
	opsStarted.Inc()

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		opsFailed.Inc()
		log.Println("Failed to select: " + err.Error())
		return status.Error(
			codes.NotFound,
			"Failed to select: "+err.Error(),
		)
	}

	for rows.Next() {
		var alb proto.Album
		err = rows.Scan(&alb.Id, &alb.Title, &alb.Artist, &alb.Score, &alb.Cover)

		if err != nil {
			opsFailed.Inc()
			log.Println("Failed to scan row: " + err.Error())
			return status.Error(
				codes.Unknown,
				"Failed to scan row: "+err.Error(),
			)
		}

		err = stream.Send(&alb)
		if err != nil {
			opsFailed.Inc()
			return status.Error(
				codes.Unknown,
				"Failed to send row: "+err.Error(),
			)
		}
	}

	err = rows.Err()
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to read row: " + err.Error())
		return status.Error(
			codes.Unknown,
			"Unable to read row: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return nil
}

// Given an UpdateRequest in, updates in.OldAlbum to be in.NewAlbum without altering the Id
func (s *Server) Update(ctx context.Context, in *proto.UpdateRequest) (*proto.Nil, error) {
	opsStarted.Inc()

	_, err := db.Exec("UPDATE album SET title=?, artist=?, price=?, cover=? WHERE id=?", in.NewAlbum.Title, in.NewAlbum.Artist, in.NewAlbum.Score, in.NewAlbum.Cover, in.OldAlbum.Id)
	if err != nil {
		opsFailed.Inc()
		log.Println("Failed to update record: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Failed to update record: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Nil{}, nil
}

// Deletes an album from the database, uses alb.Id to determine which record is deleted
func (s *Server) Delete(ctx context.Context, alb *proto.Album) (*proto.Nil, error) {
	opsStarted.Inc()

	_, err := db.Exec("DELETE FROM album WHERE id=?", alb.Id)
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to delete record: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Unable to delete record: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Nil{}, nil
}

// Starts a gRPC server for MYSQL database management
func main() {
	target := ParseEnv("TARGET_ADDRESS", ":50051")
	listener, err := net.Listen("tcp", target)
	if err != nil {
		log.Fatalln("Failed to create tcp listener", err)
	}
	defer listener.Close()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2121", nil)

	s := grpc.NewServer()
	reflection.Register(s)

	cfg := mysql.Config{
		User:   ParseEnv("MYSQL_USER", "root"),
		Passwd: ParseEnv("MYSQL_PASSWORD", "password"),
		Net:    ParseEnv("MYSQL_NETWORK_PROTOCOL", "tcp"),
		Addr:   ParseEnv("MYSQL_DATABASE_ADDRESS", "localhost:3306"),
		DBName: ParseEnv("MYSQL_DATABASE_NAME", "album"),
	}

	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalln("Failed to connect to database", err)
	}
	defer db.Close()

	proto.RegisterAlbumsServer(s, &Server{})
	err = s.Serve(listener)
	if err != nil {
		log.Fatalln("Failed to serve gRPC Server", err)
	}
}
