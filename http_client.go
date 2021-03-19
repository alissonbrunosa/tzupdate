package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Services []string

var services = Services{
	"https://ipapi.co/json",
	"http://ip-api.com/json",
	"https://freegeoip.app/json/",
	"http://worldtimeapi.org/api/ip",
}

type Client struct {
	client *http.Client
}

type response struct {
	Timezone string `json:"timezone"`
}

func (c *Client) GetTimezone(ctx context.Context, url string) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create a new request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	response := new(response)
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return nil, fmt.Errorf("unable to decode JSON response: %w", err)
	}

	return response, nil
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport,
		},
	}
}
