package watch

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestRunDetectsContentChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.md")
	if err := os.WriteFile(path, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var mu sync.Mutex
	var values []string
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, path, 10*time.Millisecond, func(data []byte) error {
			mu.Lock()
			values = append(values, string(data))
			count := len(values)
			mu.Unlock()
			if count == 1 {
				return os.WriteFile(path, []byte("second"), 0644)
			}
			cancel()
			return nil
		}, nil)
	}()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(values) != 2 || values[0] != "first" || values[1] != "second" {
		t.Fatalf("unexpected updates: %#v", values)
	}
}
