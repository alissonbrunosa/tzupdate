package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaulLocalTimePath = "/etc/localtime"
	DefaultZoneInfoPath = "/usr/share/zoneinfo"
)

func main() {
	ctx := context.Background()
	resp, err := getTimezone(ctx, services)

	if err != nil {
		fmt.Printf("Could not retrieve any timezone from the services: %v\n", err)
		os.Exit(1)
	}

	if err := setTimezone(resp.Timezone); err != nil {
		fmt.Fprintln(os.Stderr, err)

		if errors.Is(err, os.ErrPermission) {
			fmt.Fprintln(os.Stderr, "Are you root?")
		}

		os.Exit(1)
	}

	fmt.Printf("Timezone updated to %s\n", resp.Timezone)
	os.Exit(0)
}

func getTimezone(ctx context.Context, pool Services) (*response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := NewClient()

	var (
		resp    *response
		lastErr error
		mu      sync.Mutex
	)

	for _, service := range pool {
		go func(s string) {
			r, err := client.GetTimezone(ctx, s)

			switch err {
			case nil:
				cancel()
				mu.Lock()
				defer mu.Unlock()
				resp = r
			default:
				mu.Lock()
				defer mu.Unlock()
				lastErr = err
			}
		}(service)
	}

	<-ctx.Done()

	if err := ctx.Err(); err == context.DeadlineExceeded {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("cannot reach any service: %w", lastErr)
	}

	return resp, nil
}

func setTimezone(timezone string) error {
	tzPath := filepath.Join(DefaultZoneInfoPath, timezone)

	if _, err := os.Stat(tzPath); os.IsNotExist(err) {
		return fmt.Errorf("Timezone %s not supported", timezone)
	} else if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}

	if err := os.Remove(DefaulLocalTimePath); err != nil {
		return fmt.Errorf("could not remove current timezone: %w", err)
	}

	if err := os.Symlink(tzPath, DefaulLocalTimePath); err != nil {
		return fmt.Errorf("could not create symlink: %w", err)
	}

	return nil
}
