package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/ARLJohnston/go-http/proto"
	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SetupMock(t *testing.T) (mock sqlmock.Sqlmock, data *sql.DB, s *Server) {
	data, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return mock, data, &Server{db: data}
}

func TestCreateFailWhenNoDB(t *testing.T) {
	mock, _, s := SetupMock(t)

	mock.ExpectExec("INSERT INTO album").
		WithArgs("Title", "Artist", 5, "Cover").
		WillReturnError(fmt.Errorf("mock error"))

	album := proto.Album{Title: "Title", Artist: "Artist", Score: 5, Cover: "Cover"}

	id, err := s.Create(context.Background(), &album)
	if id != nil {
		t.Errorf("Expected no id to be returned, got %d", id.Id)
	}

	if err == nil {
		t.Errorf("Expected an err, got nil")
	}
	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Internal {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func TestReadFailNoDB(t *testing.T) {
	mock, data, s := SetupMock(t)
	defer data.Close()

	mock.ExpectQuery("SELECT +").
		WillReturnError(fmt.Errorf("mock error"))

	var stream proto.Albums_ReadServer

	err := s.Read(&proto.Nil{}, stream)
	if err == nil {
		t.Errorf("Expected an err, got nil")
	}

	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.NotFound {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func TestReadFailsWhenInvalidRow(t *testing.T) {
	mock, data, s := SetupMock(t)
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Name", "Row"}).
		AddRow("Invalid", "Row")

	mock.ExpectQuery("SELECT +").
		WillReturnRows(rows)

	var stream proto.Albums_ReadServer

	err := s.Read(&proto.Nil{}, stream)
	if err == nil {
		t.Errorf("Expected an err, got nil")
	}

	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Expected codes.Unknown, got %s", status.Code())
	}
}

func TestReadFailsWhenUnableToReadRow(t *testing.T) {
	mock, data, s := SetupMock(t)
	defer data.Close()

	rows := sqlmock.NewRows([]string{"id", "Title", "Artist", "Score", "Cover"}).
		AddRow(0, "Title", "Artist", 9, "cover").
		AddRow(1, "Title", "Artist", 9, "cover").
		RowError(0, fmt.Errorf("mock error")) //Need to error on first row as stream doesn't exist

	mock.ExpectQuery("SELECT +").
		WillReturnRows(rows)

	var stream proto.Albums_ReadServer

	err := s.Read(&proto.Nil{}, stream)
	if err == nil {
		t.Errorf("Expected an err, got nil")
	}

	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Expected codes.Unknown, got %s", status.Code())
	}
}

func TestUpdateFailsWhenNoDB(t *testing.T) {
	mock, data, s := SetupMock(t)
	defer data.Close()

	mock.ExpectExec("UPDATE album SET").
		WithArgs("Title", "Artist", 5, "Cover", 0).
		WillReturnError(fmt.Errorf("mock error"))

	album := proto.Album{Id: 0, Title: "Title", Artist: "Artist", Score: 5, Cover: "Cover"}

	_, err := s.Update(context.Background(), &proto.UpdateRequest{OldAlbum: &album, NewAlbum: &album})

	if err == nil {
		t.Errorf("Expected an err, got nil")
	}
	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Expected codes.Unknown got %s", status.Code())
	}
}

func TestDeleteFailsWhenNoDB(t *testing.T) {
	mock, data, s := SetupMock(t)
	defer data.Close()

	mock.ExpectExec("DELETE FROM album WHERE").
		WithArgs(0).
		WillReturnError(fmt.Errorf("mock error"))

	album := proto.Album{Id: 0, Title: "Title", Artist: "Artist", Score: 5, Cover: "Cover"}

	_, err := s.Delete(ctx, &album)
	if err == nil {
		t.Errorf("Expected an err, got nil")
	}
	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Expected codes.Unknown, got %s", status.Code())
	}
}
