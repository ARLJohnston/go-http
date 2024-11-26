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

func TestFailWhenNoDB(t *testing.T) {
	data, mock, err := sqlmock.New()
	db = data
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO album").WithArgs("Title", "Artist", 5.0, "Cover").WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

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
