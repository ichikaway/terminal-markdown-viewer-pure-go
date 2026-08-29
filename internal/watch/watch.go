package watch

import (
	"bytes"
	"context"
	"os"
	"time"
)

// Run polls path and calls onChange initially and whenever its contents change.
func Run(ctx context.Context, path string, interval time.Duration, onChange func([]byte) error, onError func(error)) error {
	previous, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := onChange(previous); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var lastError string
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			current, err := os.ReadFile(path)
			if err != nil {
				if err.Error() != lastError && onError != nil {
					onError(err)
				}
				lastError = err.Error()
				continue
			}
			lastError = ""
			if bytes.Equal(previous, current) {
				continue
			}
			previous = append(previous[:0], current...)
			if err := onChange(current); err != nil {
				return err
			}
		}
	}
}
