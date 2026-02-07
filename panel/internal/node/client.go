package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sboard/panel/internal/db"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer Doer
}

func NewClient(doer Doer) *Client {
	if doer == nil {
		doer = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{doer: doer}
}

type TrafficSample struct {
	Interface string    `json:"interface"`
	RxBytes   uint64    `json:"rx_bytes"`
	TxBytes   uint64    `json:"tx_bytes"`
	At        time.Time `json:"at"`
}

func (c *Client) Health(ctx context.Context, node db.Node) error {
	url := buildURL(node, "/api/health")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("node health status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

func (c *Client) Traffic(ctx context.Context, node db.Node) (TrafficSample, error) {
	url := buildURL(node, "/api/stats/traffic")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return TrafficSample{}, err
	}
	req.Header.Set("Authorization", "Bearer "+node.SecretKey)

	resp, err := c.doer.Do(req)
	if err != nil {
		return TrafficSample{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return TrafficSample{}, fmt.Errorf("node traffic status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var out TrafficSample
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return TrafficSample{}, err
	}
	return out, nil
}

func (c *Client) SyncConfig(ctx context.Context, node db.Node, payload any) error {
	url := buildURL(node, "/api/config/sync")
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+node.SecretKey)

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("node sync status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func buildURL(node db.Node, path string) string {
	host := strings.TrimSpace(node.APIAddress)
	if host == "" {
		host = strings.TrimSpace(node.PublicAddress)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	port := node.APIPort
	if port <= 0 {
		port = 3000
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("http://%s:%d%s", host, port, path)
}
