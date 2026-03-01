package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"sboard/panel/internal/db"
	nodev1 "sboard/panel/internal/rpc/gen/sboard/node/v1"
	nodev1connect "sboard/panel/internal/rpc/gen/sboard/node/v1/nodev1connect"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer Doer
}

var syncConfigNodeLocks = struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}{locks: map[string]*sync.Mutex{}}

type SyncError struct {
	status  int
	code    connect.Code
	summary string
}

func NewClient(doer Doer) *Client {
	if doer == nil {
		doer = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{doer: doer}
}

func (e *SyncError) Error() string {
	if e == nil {
		return ""
	}
	summary := strings.TrimSpace(e.summary)
	if summary == "" {
		summary = "rpc failed"
	}
	if e.status > 0 {
		return fmt.Sprintf("node sync status %d: %s", e.status, summary)
	}
	return "node sync request failed: " + summary
}

func (e *SyncError) HTTPStatus() int {
	if e == nil {
		return 0
	}
	return e.status
}

func (e *SyncError) Code() connect.Code {
	if e == nil {
		return connect.CodeUnknown
	}
	return e.code
}

func (e *SyncError) Retryable() bool {
	if e == nil {
		return false
	}
	return e.code == connect.CodeUnavailable || e.code == connect.CodeDeadlineExceeded
}

func (e *SyncError) Summary() string {
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.summary)
}

type TrafficSample struct {
	Interface string    `json:"interface"`
	RxBytes   uint64    `json:"rx_bytes"`
	TxBytes   uint64    `json:"tx_bytes"`
	At        time.Time `json:"at"`
}

type InboundTraffic struct {
	Tag      string    `json:"tag"`
	User     string    `json:"user"`
	Uplink   int64     `json:"uplink"`
	Downlink int64     `json:"downlink"`
	At       time.Time `json:"at"`
}

type InboundTrafficMeta struct {
	TrackedTags int   `json:"tracked_tags"`
	TCPConns    int64 `json:"tcp_conns"`
	UDPConns    int64 `json:"udp_conns"`
}

func (c *Client) Health(ctx context.Context, node db.Node) error {
	rpcClient := c.newRPCClient(node)
	_, err := rpcClient.Health(ctx, &nodev1.HealthRequest{})
	if err == nil {
		return nil
	}
	return formatRPCError("node health", err)
}

func (c *Client) Traffic(ctx context.Context, node db.Node) (TrafficSample, error) {
	rpcClient := c.newRPCClient(node)
	out, err := rpcClient.GetTraffic(ctx, &nodev1.GetTrafficRequest{})
	if err == nil {
		parsedAt, parseErr := parseRFC3339(out.GetAt())
		if parseErr != nil {
			return TrafficSample{}, parseErr
		}
		return TrafficSample{
			Interface: out.GetInterface(),
			RxBytes:   out.GetRxBytes(),
			TxBytes:   out.GetTxBytes(),
			At:        parsedAt,
		}, nil
	}
	return TrafficSample{}, formatRPCError("node traffic", err)
}

func (c *Client) InboundTraffic(ctx context.Context, node db.Node, reset bool) ([]InboundTraffic, error) {
	rows, _, err := c.InboundTrafficWithMeta(ctx, node, reset)
	return rows, err
}

func (c *Client) InboundTrafficWithMeta(ctx context.Context, node db.Node, reset bool) ([]InboundTraffic, *InboundTrafficMeta, error) {
	rpcClient := c.newRPCClient(node)
	out, err := rpcClient.GetInboundTraffic(ctx, &nodev1.GetInboundTrafficRequest{Reset_: reset})
	if err == nil {
		rows := make([]InboundTraffic, 0, len(out.GetData()))
		for _, item := range out.GetData() {
			parsedAt, parseErr := parseRFC3339(item.GetAt())
			if parseErr != nil {
				return nil, nil, parseErr
			}
			rows = append(rows, InboundTraffic{
				Tag:      item.GetTag(),
				User:     item.GetUser(),
				Uplink:   item.GetUplink(),
				Downlink: item.GetDownlink(),
				At:       parsedAt,
			})
		}

		var meta *InboundTrafficMeta
		if out.GetMeta() != nil {
			meta = &InboundTrafficMeta{
				TrackedTags: int(out.GetMeta().GetTrackedTags()),
				TCPConns:    out.GetMeta().GetTcpConns(),
				UDPConns:    out.GetMeta().GetUdpConns(),
			}
		}
		return rows, meta, nil
	}
	return nil, nil, formatRPCError("node inbound traffic", err)
}

func (c *Client) SyncConfig(ctx context.Context, node db.Node, payload any) error {
	lock := syncConfigNodeLock(node)
	lock.Lock()
	defer lock.Unlock()

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	rpcClient := c.newRPCClient(node)
	_, err = rpcClient.SyncConfig(ctx, &nodev1.SyncConfigRequest{PayloadJson: b})
	if err == nil {
		return nil
	}
	return formatSyncRPCError(err)
}

func (c *Client) newRPCClient(node db.Node) nodev1connect.NodeControlServiceClient {
	secret := strings.TrimSpace(node.SecretKey)
	authInterceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if secret != "" {
				req.Header().Set("Authorization", "Bearer "+secret)
			}
			return next(ctx, req)
		}
	})
	return nodev1connect.NewNodeControlServiceClient(c.doer, buildRPCBaseURL(node), connect.WithInterceptors(authInterceptor))
}

func parseRFC3339(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func rpcCodeToHTTPStatus(code connect.Code) int {
	switch code {
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest
	case connect.CodeUnauthenticated:
		return http.StatusUnauthorized
	case connect.CodeNotFound:
		return http.StatusNotFound
	case connect.CodeResourceExhausted:
		return http.StatusRequestEntityTooLarge
	case connect.CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case connect.CodeUnavailable:
		return http.StatusBadGateway
	case connect.CodeUnimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

func formatRPCError(prefix string, err error) error {
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "rpc failed"
	}

	var connErr *connect.Error
	if errors.As(err, &connErr) {
		status := rpcCodeToHTTPStatus(connErr.Code())
		if status > 0 {
			return fmt.Errorf("%s status %d: %s", prefix, status, msg)
		}
	}
	return fmt.Errorf("%s request failed: %s", prefix, msg)
}

func formatSyncRPCError(err error) error {
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		msg = "rpc failed"
	}

	var connErr *connect.Error
	if errors.As(err, &connErr) {
		return &SyncError{
			status:  rpcCodeToHTTPStatus(connErr.Code()),
			code:    connErr.Code(),
			summary: msg,
		}
	}
	return &SyncError{summary: msg}
}

func syncConfigNodeLock(node db.Node) *sync.Mutex {
	key := buildRPCBaseURL(node)
	syncConfigNodeLocks.mu.Lock()
	defer syncConfigNodeLocks.mu.Unlock()
	if lock, ok := syncConfigNodeLocks.locks[key]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	syncConfigNodeLocks.locks[key] = lock
	return lock
}

func buildRPCBaseURL(node db.Node) string {
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
	return fmt.Sprintf("http://%s:%d/rpc", host, port)
}
