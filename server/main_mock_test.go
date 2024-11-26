package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ARLJohnston/go-http/pb"
	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateFailWhenNoDB(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO album").WithArgs("Title", "Artist", 5.0, "Cover").WillReturnError(fmt.Errorf("some error"))

	album := pb.Album{Title: "Title", Artist: "Artist", Price: 5.0, Cover: "Cover"}

	s := &Server{}
	ctx := context.Background()

	id, err := s.Create(ctx, &album)
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

func TestCreateFailWhenNoIdentifier(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO album").WithArgs("Title", "Artist", 5.0, "Cover").WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("some error")))

	s := &Server{}
	ctx := context.Background()

	album := pb.Album{Title: "Title", Artist: "Artist", Price: 5.0, Cover: "Cover"}

	id, err := s.Create(ctx, &album)
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
	if status.Code() != codes.NotFound {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func TestReadFailNoDB(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := &Server{}

	mock.ExpectQuery("SELECT * FROM album").WillReturnError(fmt.Errorf("some error"))

	var stream pb.Albums_ReadServer

	err = s.Read(&pb.Nil{}, stream)
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

func TestUpdateFailsWhenNoDB(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE album SET").WithArgs("Title", "Artist", 5.0, "Cover", 0).WillReturnError(fmt.Errorf("some error"))

	album := pb.Album{ID: 0, Title: "Title", Artist: "Artist", Price: 5.0, Cover: "Cover"}

	s := &Server{}
	ctx := context.Background()

	_, err = s.Update(ctx, &pb.UpdateRequest{OldAlbum: &album, NewAlbum: &album})

	if err == nil {
		t.Errorf("Expected an err, got nil")
	}
	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func TestDelereFailsWhenNoDB(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM album WHERE").WithArgs(0).WillReturnError(fmt.Errorf("some error"))

	album := pb.Album{ID: 0, Title: "Title", Artist: "Artist", Price: 5.0, Cover: "Cover"}

	s := &Server{}
	ctx := context.Background()

	_, err = s.Delete(ctx, &album)
	if err == nil {
		t.Errorf("Expected an err, got nil")
	}
	status, ok := status.FromError(err)
	if !ok {
		t.Errorf("Unable to convert error to status")
	}
	if status.Code() != codes.Unknown {
		t.Errorf("Unexpected error returned: %v", err)
	}
}
