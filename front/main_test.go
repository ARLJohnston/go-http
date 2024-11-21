package main

import (
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
