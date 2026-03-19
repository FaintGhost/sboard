package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config holds the settings needed by the heartbeat loop.
type Config struct {
	PanelURL  string
	NodeUUID  string
	SecretKey string
	APIAddr   string
	Version   string
	Interval  time.Duration
}

type heartbeatRequest struct {
	UUID      string `json:"uuid"`
	SecretKey string `json:"secretKey"`
	Version   string `json:"version,omitempty"`
	APIAddr   string `json:"apiAddr,omitempty"`
}

type heartbeatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

const heartbeatPath = "/sboard.panel.v1.NodeRegistrationService/Heartbeat"

// Run starts the heartbeat loop. It blocks until ctx is cancelled.
// If cfg.PanelURL is empty, it returns immediately (heartbeat disabled).
// If client is nil, a default HTTP client with a 10s timeout is used.
func Run(ctx context.Context, cfg Config, client *http.Client) {
	if cfg.PanelURL == "" {
		return
	}

	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	log.Printf("[heartbeat] starting: panel=%s uuid=%s interval=%s", cfg.PanelURL, cfg.NodeUUID, cfg.Interval)

	// Send immediately on startup, then on every tick.
	if err := sendHeartbeat(ctx, client, cfg); err != nil {
		log.Printf("[heartbeat] error: %v", err)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[heartbeat] stopped")
			return
		case <-ticker.C:
			if err := sendHeartbeat(ctx, client, cfg); err != nil {
				log.Printf("[heartbeat] error: %v", err)
			}
		}
	}
}

func sendHeartbeat(ctx context.Context, client *http.Client, cfg Config) error {
	reqBody := heartbeatRequest{
		UUID:      cfg.NodeUUID,
		SecretKey: cfg.SecretKey,
		Version:   cfg.Version,
		APIAddr:   cfg.APIAddr,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url, err := buildHeartbeatURL(cfg.PanelURL)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var hbResp heartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&hbResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	switch hbResp.Status {
	case "NODE_HEARTBEAT_STATUS_RECOGNIZED":
		log.Printf("[heartbeat] recognized by panel")
	case "NODE_HEARTBEAT_STATUS_PENDING":
		log.Printf("[heartbeat] pending approval on panel")
	case "NODE_HEARTBEAT_STATUS_REJECTED":
		log.Printf("[heartbeat] rejected by panel")
	default:
		log.Printf("[heartbeat] unknown status: %s", hbResp.Status)
	}

	return nil
}

func buildHeartbeatURL(rawBase string) (string, error) {
	base, err := normalizePanelRPCBaseURL(rawBase)
	if err != nil {
		return "", err
	}
	return base + heartbeatPath, nil
}

func normalizePanelRPCBaseURL(rawBase string) (string, error) {
	trimmed := strings.TrimSpace(rawBase)
	if trimmed == "" {
		return "", fmt.Errorf("panel url is empty")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse panel url: %w", err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid panel url: missing host")
	}

	path := strings.TrimRight(parsed.Path, "/")
	if path == "" {
		path = "/rpc"
	} else if !strings.HasSuffix(path, "/rpc") {
		path += "/rpc"
	}
	parsed.Path = path
	parsed.RawPath = ""

	return strings.TrimRight(parsed.String(), "/"), nil
}
