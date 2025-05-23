// Package main implements a gRPC server for interaction with a postgres database
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

	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ARLJohnston/go-http/proto"
)

var (
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
	powerUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "estimated_power_usage_W",
		Help: "Estimated power usage for a Pi3B running this application",
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
	db *sql.DB // Handle to the database
	proto.UnimplementedAlbumsServer
}

// Creates an Album alb in the db, returns the SQL identifier for that album in the database
func (s *Server) Create(ctx context.Context, alb *proto.Album) (*proto.Identifier, error) {
	opsStarted.Inc()

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return nil, status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	_, err := s.db.Exec("INSERT INTO album (title, artist, score, cover) VALUES ($1, $2, $3, $4)", alb.Title, alb.Artist, alb.Score, alb.Cover)
	if err != nil {
		opsFailed.Inc()
		log.Println("Create failed:" + err.Error())
		return nil, status.Error(
			codes.Internal, "Create failed: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Identifier{Id: 1}, nil
}

// Opens stream for a streaming read of every album in the database
func (s *Server) Read(_ *proto.Nil, stream proto.Albums_ReadServer) error {
	opsStarted.Inc()

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	rows, err := s.db.Query("SELECT * FROM album FETCH FIRST 16 ROWS ONLY")
	if err != nil {
		opsFailed.Inc()
		log.Println("Failed to select: " + err.Error())
		return status.Error(
			codes.NotFound,
			"Failed to select: "+err.Error(),
		)
	}
	defer rows.Close()

	var count uint

	for rows.Next() && count <= 16 {
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
		count++
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

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return nil, status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	_, err := s.db.Exec("UPDATE album SET title=$1, artist=$2, score=$3, cover=$4 WHERE id=$5", in.NewAlbum.Title, in.NewAlbum.Artist, in.NewAlbum.Score, in.NewAlbum.Cover, in.OldAlbum.Id)
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

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return nil, status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	_, err := s.db.Exec("DELETE FROM album WHERE id=$1", alb.Id)
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

func (s *Server) Increment(ctx context.Context, in *proto.Identifier) (*proto.Score, error) {
	opsStarted.Inc()

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return nil, status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	_, err := s.db.Exec("UPDATE album SET score = score + 1 WHERE id=$1", in.Id)
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to increment score: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Unable to increment score: "+err.Error(),
		)
	}

	var score int
	err = s.db.QueryRow("SELECT score FROM album WHERE id=$1", in.Id).Scan(&score)
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to retrieve score: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Unable to retrieve score: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Score{Score: int64(score)}, nil
}

func (s *Server) Decrement(ctx context.Context, in *proto.Identifier) (*proto.Score, error) {
	opsStarted.Inc()

	if s.db == nil {
		opsFailed.Inc()
		log.Println("Unable to connect to database")
		return nil, status.Error(
			codes.NotFound, "Unable to connect to database",
		)
	}

	_, err := s.db.Exec("UPDATE album SET score = score - 1 WHERE id=$1", in.Id)
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to decrement score: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Unable to decrement score: "+err.Error(),
		)
	}

	var score int
	err = s.db.QueryRow("SELECT score FROM album WHERE id=$1", in.Id).Scan(&score)
	if err != nil {
		opsFailed.Inc()
		log.Println("Unable to retrieve score: " + err.Error())
		return nil, status.Error(
			codes.Unknown,
			"Unable to retrieve score: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &proto.Score{Score: int64(score)}, nil
}

// Starts a gRPC server for database management
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

	var server Server

	host := ParseEnv("DATABASE_ADDRESS", "localhost")
	connStr := "postgres://user:password@" + host + "/album?sslmode=disable"
	server.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("Failed to connect to database: ", err)
	}
	defer server.db.Close()
	server.db.SetMaxOpenConns(10)
	server.db.SetMaxIdleConns(5)

	if err := server.db.Ping(); err != nil {
		log.Fatalln("Unable to ping database: ", err)
	}

	proto.RegisterAlbumsServer(s, &server)
	err = s.Serve(listener)
	if err != nil {
		log.Fatalln("Failed to serve gRPC Server", err)
	}
}
