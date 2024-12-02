package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseEnv(t *testing.T) {
	os.Setenv("VAR", "variable")
	got := parseEnv("VAR", "fallback")
	want := "variable"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}

}

func TestParseEnvFallback(t *testing.T) {
	got := parseEnv("veryspecific", "fallback")
	want := "fallback"

	if got != want {
		t.Errorf("got %s wanted %s", got, want)
	}
}

func TestPagePosts(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handleLoad(wr, req)
	if wr.Code != http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}
}
