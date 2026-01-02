package main

import (
	"context"
	"testing"
)

// TODO: Refactor url_checker to allow mocking IO operations

func TestFileNotFound(t *testing.T) {
	ctx := context.Background()

	err := run(ctx, "test.txt", 5)

	if err == nil || err.Error() != "open test.txt: no such file or directory" {
		t.Errorf("ERROR: %v\n", err)
	}
}

func TestUrls(t *testing.T) {
	ctx := context.Background()

	err := run(ctx, "urls.txt", 5)

	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
}
