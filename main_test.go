package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {
	t.Run("Test http", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		w := httptest.NewRecorder()

		hello(w, req)
		res := w.Result()
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if string(data) != "Hello, World!" {
			t.Errorf("got %s want %s", string(data), "Hello, World!")
		}
	})

}
