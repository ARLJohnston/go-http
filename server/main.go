package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/ARLJohnston/go-http/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"database/sql"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

var cfg mysql.Config

type server struct {
	pb.UnimplementedAlbumsServer
}

func (s *server) Create(ctx context.Context, alb *pb.Album) (*pb.Identifier, error) {
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO album (title, artist, price, cover) VALUES (?, ?, ?, ?)", alb.Title, alb.Artist, alb.Price, alb.Cover)
	if err != nil {
		return nil, status.Error(
			codes.Unknown, "Create failed: "+err.Error(),
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, status.Error(
			codes.NotFound, "Failed to get last inset id: "+err.Error(),
		)
	}

	return &pb.Identifier{Id: id}, nil
}

func (s *server) Read(_ *pb.Nil, stream pb.Albums_ReadServer) error {
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		return status.Error(
			codes.Unknown,
			"Failed to select: "+err.Error(),
		)
	}

	for rows.Next() {
		var alb pb.Album
		err = rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price, &alb.Cover)
		if err != nil {
			return status.Error(
				codes.Unknown,
				"Failed to scan row: "+err.Error(),
			)
		}

		err = stream.Send(&alb)
		if err != nil {
			return status.Error(
				codes.Unknown,
				"Failed to send row: "+err.Error(),
			)
		}
	}

	if err := rows.Err(); err != nil {
		return status.Error(
			codes.Unknown,
			"Unable to read row: "+err.Error(),
		)
	}

	return nil
}

func (s *server) Update(ctx context.Context, in *pb.UpdateRequest) (*pb.Nil, error) {
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE album SET title=?, artist=?, price=?, cover=? WHERE id=?", in.NewAlbum.Title, in.NewAlbum.Artist, in.NewAlbum.Price, in.NewAlbum.Cover, in.OldAlbum.ID)
	if err != nil {
		return nil, status.Error(
			codes.Unknown,
			"Failed to update record: "+err.Error(),
		)
	}

	return &pb.Nil{}, nil
}

func (s *server) Delete(ctx context.Context, alb *pb.Album) (*pb.Nil, error) {
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM album WHERE id=?", alb.ID)
	if err != nil {
		return nil, status.Error(
			codes.Unknown,
			"Unable to delete record: "+err.Error(),
		)
	}

	return &pb.Nil{}, nil
}

func parseEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func main() {
	cfg = mysql.Config{
		User:   parseEnv("MYSQL_USER", "dbuser"),
		Passwd: parseEnv("MYSQL_USER_PASSWORD", "userpass"),
		Net:    parseEnv("MYSQL_NETWORK_PROTOCOL", "tcp"),
		Addr:   parseEnv("MYSQL_DATABASE_ADDRESS", "localhost:3306"),
		DBName: parseEnv("MYSQL_DATABASE_NAME", "album"),
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to create tcp listener", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	pb.RegisterAlbumsServer(s, &server{})
	err = s.Serve(listener)
	if err != nil {
		log.Fatalln("Failed to serve gRPC Server", err)
	}
}
