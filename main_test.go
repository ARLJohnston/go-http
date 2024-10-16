package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	w := httptest.NewRecorder()

	t.Run("Put file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/data/example", nil)

		put(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if string(data) != "File was created" {
			t.Errorf("got %s want %s", string(data), "File was created")
		}
	})

	t.Run("Get file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data/example", nil)

		get(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if string(data) != "Example text" {
			t.Errorf("got %s want %s", string(data), "Example text")
		}
	})

}
