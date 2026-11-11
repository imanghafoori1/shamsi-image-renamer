package main

import (
	"errors"
	"testing"
)

func TestDoneMessage(t *testing.T) {
	if got := doneMessage(5, 2); got != "Done. Renamed 5, skipped 2." {
		t.Errorf("got %q", got)
	}
}

func TestDoneTaggedMessage(t *testing.T) {
	if got := doneTaggedMessage(5, 2, 3); got != "Done. Renamed 5, skipped 2, tagged 3." {
		t.Errorf("got %q", got)
	}
}

func TestProgressMessage(t *testing.T) {
	if got := progressMessage(20, 100, "IMG_0012.jpg"); got != "Processing 20/100: IMG_0012.jpg" {
		t.Errorf("got %q", got)
	}
}

func TestStartingMessage(t *testing.T) {
	if got := startingMessage(); got != "Starting..." {
		t.Errorf("got %q", got)
	}
}

func TestErrorPathMessage(t *testing.T) {
	if got := errorPathMessage(); got != "Please enter a Path" {
		t.Errorf("got %q", got)
	}
}

func TestErrorPrefix(t *testing.T) {
	err := errors.New("boom")
	if got := errorPrefix(err.Error()); got != "Error: boom" {
		t.Errorf("got %q", got)
	}
}
