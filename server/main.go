package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"database/sql"

	"github.com/go-sql-driver/mysql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/ARLJohnston/go-http/pb"
)

var (
	db *sql.DB

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

func parseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

type server struct {
	pb.UnimplementedAlbumsServer
	cfg mysql.Config
}

func (s *server) Create(ctx context.Context, alb *pb.Album) (*pb.Identifier, error) {
	opsStarted.Inc()
	var err error
	db, err = sql.Open("mysql", s.cfg.FormatDSN())
	if err != nil {
		opsFailed.Inc()
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO album (title, artist, price, cover) VALUES (?, ?, ?, ?)", alb.Title, alb.Artist, alb.Price, alb.Cover)
	if err != nil {
		opsFailed.Inc()
		return nil, status.Error(
			codes.Unknown, "Create failed: "+err.Error(),
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		opsFailed.Inc()
		return nil, status.Error(
			codes.NotFound, "Failed to get last inset id: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &pb.Identifier{Id: id}, nil
}

func (s *server) Read(_ *pb.Nil, stream pb.Albums_ReadServer) error {
	opsStarted.Inc()
	var err error
	db, err = sql.Open("mysql", s.cfg.FormatDSN())
	if err != nil {
		opsFailed.Inc()
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		opsFailed.Inc()
		return status.Error(
			codes.Unknown,
			"Failed to select: "+err.Error(),
		)
	}

	for rows.Next() {
		var alb pb.Album
		err = rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price, &alb.Cover)
		if err != nil {
			opsFailed.Inc()
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

	if err := rows.Err(); err != nil {
		opsFailed.Inc()
		return status.Error(
			codes.Unknown,
			"Unable to read row: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return nil
}

func (s *server) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.Nil, error) {
	opsStarted.Inc()
	var err error
	db, err = sql.Open("mysql", s.cfg.FormatDSN())
	if err != nil {
		opsFailed.Inc()
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE album SET title=?, artist=?, price=?, cover=? WHERE id=?", in.NewAlbum.Title, in.NewAlbum.Artist, in.NewAlbum.Price, in.NewAlbum.Cover, in.OldAlbum.ID)
	if err != nil {
		opsFailed.Inc()
		return nil, status.Error(
			codes.Unknown,
			"Failed to update record: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &pb.Nil{}, nil
}

func (s *server) Delete(ctx context.Context, alb *pb.Album) (*pb.Nil, error) {
	opsStarted.Inc()
	var err error
	db, err = sql.Open("mysql", s.cfg.FormatDSN())
	if err != nil {
		opsFailed.Inc()
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM album WHERE id=?", alb.ID)
	if err != nil {
		opsFailed.Inc()
		return nil, status.Error(
			codes.Unknown,
			"Unable to delete record: "+err.Error(),
		)
	}

	opsSucceeded.Inc()
	return &pb.Nil{}, nil
}

func main() {
	target := parseEnv("TARGET_ADDRESS", ":50051")
	listener, err := net.Listen("tcp", target)
	if err != nil {
		log.Fatalln("Failed to create tcp listener", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	cfg := mysql.Config{
		User:   parseEnv("MYSQL_USER", "root"),
		Passwd: parseEnv("MYSQL_PASSWORD", "password"),
		Net:    parseEnv("MYSQL_NETWORK_PROTOCOL", "tcp"),
		Addr:   parseEnv("MYSQL_DATABASE_ADDRESS", "localhost:3306"),
		DBName: parseEnv("MYSQL_DATABASE_NAME", "album"),
	}

	pb.RegisterAlbumsServer(s, &server{cfg: cfg})
	err = s.Serve(listener)
	if err != nil {
		log.Fatalln("Failed to serve gRPC Server", err)
	}
}
