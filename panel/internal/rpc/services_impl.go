package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"sboard/panel/internal/api"
	"sboard/panel/internal/buildinfo"
	"sboard/panel/internal/db"
	inbval "sboard/panel/internal/inbounds"
	"sboard/panel/internal/node"
	"sboard/panel/internal/password"
	panelv1 "sboard/panel/internal/rpc/gen/sboard/panel/v1"
	"sboard/panel/internal/singboxcli"
	"sboard/panel/internal/traffic"
	"sboard/panel/internal/userstate"
)

func unexpectedResponseError() error {
	return connect.NewError(connect.CodeInternal, errors.New("unexpected response"))
}

func encodeJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func decodeJSONMap(s string) map[string]any {
	m := map[string]any{}
	trim := s
	if trim == "" {
		return m
	}
	if err := json.Unmarshal([]byte(trim), &m); err != nil {
		return map[string]any{}
	}
	return m
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := formatTime(*t)
	return &v
}

type syncResult = api.SyncResult

type inboundDTO = api.Inbound
type syncJobListItemDTO = api.SyncJobListItem
type syncAttemptDTO = api.SyncAttempt

func mapSyncResult(r syncResult) *panelv1.SyncResult {
	out := &panelv1.SyncResult{Status: r.Status}
	if r.Error != nil {
		out.Error = r.Error
	}
	return out
}

func rawJSONToString(raw json.RawMessage) string {
	if len(raw) == 0 || !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

func mapInbound(inb inboundDTO) *panelv1.Inbound {
	return &panelv1.Inbound{
		Id:                    inb.Id,
		Uuid:                  inb.Uuid,
		NodeId:                inb.NodeId,
		Tag:                   inb.Tag,
		Protocol:              inb.Protocol,
		ListenPort:            int32(inb.ListenPort),
		PublicPort:            int32(inb.PublicPort),
		SettingsJson:          encodeJSON(inb.Settings),
		TlsSettingsJson:       encodeJSON(inb.TlsSettings),
		TransportSettingsJson: encodeJSON(inb.TransportSettings),
	}
}

func mapDBInbound(inb db.Inbound) *panelv1.Inbound {
	return &panelv1.Inbound{
		Id:                    inb.ID,
		Uuid:                  inb.UUID,
		NodeId:                inb.NodeID,
		Tag:                   inb.Tag,
		Protocol:              inb.Protocol,
		ListenPort:            int32(inb.ListenPort),
		PublicPort:            int32(inb.PublicPort),
		SettingsJson:          rawJSONToString(inb.Settings),
		TlsSettingsJson:       rawJSONToString(inb.TLSSettings),
		TransportSettingsJson: rawJSONToString(inb.TransportSettings),
	}
}

func mapSyncJob(item syncJobListItemDTO) *panelv1.SyncJobListItem {
	return &panelv1.SyncJobListItem{
		Id:              item.Id,
		NodeId:          item.NodeId,
		ParentJobId:     item.ParentJobId,
		TriggerSource:   item.TriggerSource,
		Status:          item.Status,
		InboundCount:    int32(item.InboundCount),
		ActiveUserCount: int32(item.ActiveUserCount),
		PayloadHash:     item.PayloadHash,
		AttemptCount:    int32(item.AttemptCount),
		DurationMs:      item.DurationMs,
		ErrorSummary:    item.ErrorSummary,
		CreatedAt:       formatTime(item.CreatedAt),
		StartedAt:       formatTimePtr(item.StartedAt),
		FinishedAt:      formatTimePtr(item.FinishedAt),
	}
}

func mapDBSyncJob(item db.SyncJob) *panelv1.SyncJobListItem {
	return &panelv1.SyncJobListItem{
		Id:              item.ID,
		NodeId:          item.NodeID,
		ParentJobId:     item.ParentJobID,
		TriggerSource:   item.TriggerSource,
		Status:          item.Status,
		InboundCount:    int32(item.InboundCount),
		ActiveUserCount: int32(item.ActiveUserCount),
		PayloadHash:     item.PayloadHash,
		AttemptCount:    int32(item.AttemptCount),
		DurationMs:      item.DurationMS,
		ErrorSummary:    item.ErrorSummary,
		CreatedAt:       formatTime(item.CreatedAt),
		StartedAt:       formatTimePtr(item.StartedAt),
		FinishedAt:      formatTimePtr(item.FinishedAt),
	}
}

func mapSyncAttempt(item syncAttemptDTO) *panelv1.SyncAttempt {
	return &panelv1.SyncAttempt{
		Id:           item.Id,
		AttemptNo:    int32(item.AttemptNo),
		Status:       item.Status,
		HttpStatus:   int32(item.HttpStatus),
		DurationMs:   item.DurationMs,
		ErrorSummary: item.ErrorSummary,
		BackoffMs:    item.BackoffMs,
		StartedAt:    formatTime(item.StartedAt),
		FinishedAt:   formatTimePtr(item.FinishedAt),
	}
}

func mapDBSyncAttempt(item db.SyncAttempt) *panelv1.SyncAttempt {
	return &panelv1.SyncAttempt{
		Id:           item.ID,
		AttemptNo:    int32(item.AttemptNo),
		Status:       item.Status,
		HttpStatus:   int32(item.HTTPStatus),
		DurationMs:   item.DurationMS,
		ErrorSummary: item.ErrorSummary,
		BackoffMs:    item.BackoffMS,
		StartedAt:    formatTime(item.StartedAt),
		FinishedAt:   formatTimePtr(item.FinishedAt),
	}
}

func mapDBUser(user db.User, groupIDs []int64) *panelv1.User {
	out := &panelv1.User{
		Id:              user.ID,
		Uuid:            user.UUID,
		Username:        user.Username,
		GroupIds:        append([]int64{}, groupIDs...),
		TrafficLimit:    user.TrafficLimit,
		TrafficUsed:     user.TrafficUsed,
		TrafficResetDay: int32(user.TrafficResetDay),
		Status:          userstate.EffectiveStatus(user, time.Now().UTC()),
	}
	if user.ExpireAt != nil {
		v := user.ExpireAt.In(time.Local).Format(time.RFC3339)
		out.ExpireAt = &v
	}
	return out
}

func mapDBGroup(group db.Group) *panelv1.Group {
	return &panelv1.Group{
		Id:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		MemberCount: group.MemberCount,
	}
}

func mapDBNode(node db.Node) *panelv1.Node {
	out := &panelv1.Node{
		Id:            node.ID,
		Uuid:          node.UUID,
		Name:          node.Name,
		ApiAddress:    node.APIAddress,
		ApiPort:       int32(node.APIPort),
		SecretKey:     node.SecretKey,
		PublicAddress: node.PublicAddress,
		Status:        node.Status,
		GroupId:       node.GroupID,
	}
	if node.LastSeenAt != nil {
		v := node.LastSeenAt.In(time.Local).Format(time.RFC3339)
		out.LastSeenAt = &v
	}
	return out
}

func parseTimeRFC3339Ptr(v *string) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid time format"))
	}
	return &parsed, nil
}

func intPtrFromInt32Ptr(v *int32) *int {
	if v == nil {
		return nil
	}
	n := int(*v)
	return &n
}

func nonEmptyOrNA(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return "N/A"
	}
	return v
}

func signAdminToken(secret string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	claims := jwt.RegisteredClaims{
		Subject:   "admin",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func normalizeSubscriptionBaseURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid subscription_base_url")
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", errors.New("subscription_base_url must use http or https")
	}

	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("subscription_base_url must be protocol + ip:port")
	}

	host := strings.TrimSpace(parsed.Hostname())
	if net.ParseIP(host) == nil {
		return "", errors.New("subscription_base_url must use a valid IP")
	}

	portStr := strings.TrimSpace(parsed.Port())
	if portStr == "" {
		return "", errors.New("subscription_base_url must include port")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", errors.New("subscription_base_url has invalid port")
	}

	return scheme + "://" + net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func normalizeTimezone(raw string) (string, *time.Location, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = "UTC"
	}
	loc, err := time.LoadLocation(value)
	if err != nil {
		return "", nil, errors.New("invalid timezone")
	}
	return value, loc, nil
}

func currentTimezoneName() string {
	name := strings.TrimSpace(time.Local.String())
	if name == "" {
		return "UTC"
	}
	return name
}

func (s *Server) GetHealth(ctx context.Context, _ *panelv1.GetHealthRequest) (*panelv1.GetHealthResponse, error) {
	return &panelv1.GetHealthResponse{Status: "ok"}, nil
}

func (s *Server) GetBootstrapStatus(ctx context.Context, _ *panelv1.GetBootstrapStatusRequest) (*panelv1.GetBootstrapStatusResponse, error) {
	n, err := db.AdminCount(s.store)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	return &panelv1.GetBootstrapStatusResponse{Data: &panelv1.BootstrapStatus{NeedsSetup: n == 0}}, nil
}

func (s *Server) Bootstrap(ctx context.Context, req *panelv1.BootstrapRequest) (*panelv1.BootstrapResponse, error) {
	setupToken := strings.TrimSpace(req.GetSetupToken())
	if setupToken == "" && req.XSetupToken_ != nil {
		setupToken = strings.TrimSpace(*req.XSetupToken_)
	}
	if strings.TrimSpace(s.cfg.SetupToken) == "" || setupToken != strings.TrimSpace(s.cfg.SetupToken) {
		return nil, connectErrorFromHTTP(401, "unauthorized")
	}

	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, connectErrorFromHTTP(400, "missing username")
	}
	if strings.TrimSpace(req.GetPassword()) == "" {
		return nil, connectErrorFromHTTP(400, "missing password")
	}
	if req.GetPassword() != req.GetConfirmPassword() {
		return nil, connectErrorFromHTTP(400, "passwords do not match")
	}
	if len(req.GetPassword()) < 8 {
		return nil, connectErrorFromHTTP(400, "password too short")
	}

	h, err := password.Hash(req.GetPassword())
	if err != nil {
		return nil, connectErrorFromHTTP(400, "invalid password")
	}
	created, err := db.AdminCreateIfNone(s.store, username, h)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	if !created {
		return nil, connectErrorFromHTTP(409, "already initialized")
	}

	return &panelv1.BootstrapResponse{Data: &panelv1.BootstrapResult{Ok: true}}, nil
}

func (s *Server) Login(ctx context.Context, req *panelv1.LoginRequest) (*panelv1.LoginResponseEnvelope, error) {
	n, err := db.AdminCount(s.store)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	if n == 0 {
		return nil, connectErrorFromHTTP(428, "needs setup")
	}

	a, ok, err := db.AdminGetByUsername(s.store, req.GetUsername())
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	if !ok || !password.Verify(a.PasswordHash, req.GetPassword()) {
		return nil, connectErrorFromHTTP(401, "unauthorized")
	}

	token, exp, err := signAdminToken(s.cfg.JWTSecret)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "sign token failed")
	}
	return &panelv1.LoginResponseEnvelope{Data: &panelv1.LoginResult{Token: token, ExpiresAt: formatTime(exp)}}, nil
}

func (s *Server) GetAdminProfile(ctx context.Context, _ *panelv1.GetAdminProfileRequest) (*panelv1.GetAdminProfileResponse, error) {
	admin, ok, err := db.AdminGetFirst(s.store)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	if !ok {
		return nil, connectErrorFromHTTP(428, "needs setup")
	}
	return &panelv1.GetAdminProfileResponse{Data: &panelv1.AdminProfile{Username: admin.Username}}, nil
}

func (s *Server) UpdateAdminProfile(ctx context.Context, req *panelv1.UpdateAdminProfileRequest) (*panelv1.UpdateAdminProfileResponse, error) {
	admin, ok, err := db.AdminGetFirst(s.store)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "db error")
	}
	if !ok {
		return nil, connectErrorFromHTTP(428, "needs setup")
	}

	if strings.TrimSpace(req.GetOldPassword()) == "" {
		return nil, connectErrorFromHTTP(400, "missing old_password")
	}
	if !password.Verify(admin.PasswordHash, req.GetOldPassword()) {
		return nil, connectErrorFromHTTP(401, "old password mismatch")
	}

	newUsername := ""
	if req.NewUsername != nil {
		newUsername = strings.TrimSpace(*req.NewUsername)
	}
	if newUsername == "" {
		newUsername = admin.Username
	}

	newPasswordHash := admin.PasswordHash
	hasPasswordChange := (req.NewPassword != nil && strings.TrimSpace(*req.NewPassword) != "") ||
		(req.ConfirmPassword != nil && strings.TrimSpace(*req.ConfirmPassword) != "")
	if hasPasswordChange {
		newPw := ""
		confirmPw := ""
		if req.NewPassword != nil {
			newPw = *req.NewPassword
		}
		if req.ConfirmPassword != nil {
			confirmPw = *req.ConfirmPassword
		}
		if newPw != confirmPw {
			return nil, connectErrorFromHTTP(400, "passwords do not match")
		}
		if len(newPw) < 8 {
			return nil, connectErrorFromHTTP(400, "password too short")
		}
		hashed, err := password.Hash(newPw)
		if err != nil {
			return nil, connectErrorFromHTTP(400, "invalid password")
		}
		newPasswordHash = hashed
	}

	if newUsername == admin.Username && newPasswordHash == admin.PasswordHash {
		return nil, connectErrorFromHTTP(400, "no changes")
	}

	if err := db.AdminUpdateCredentials(s.store, admin.ID, newUsername, newPasswordHash); err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "username exists")
		}
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "admin not found")
		}
		return nil, connectErrorFromHTTP(500, "db error")
	}

	return &panelv1.UpdateAdminProfileResponse{Data: &panelv1.AdminProfile{Username: newUsername}}, nil
}

func (s *Server) GetSystemInfo(ctx context.Context, _ *panelv1.GetSystemInfoRequest) (*panelv1.GetSystemInfoResponse, error) {
	return &panelv1.GetSystemInfoResponse{Data: &panelv1.SystemInfo{
		PanelVersion:   nonEmptyOrNA(buildinfo.PanelVersion),
		PanelCommitId:  nonEmptyOrNA(buildinfo.PanelCommitID),
		SingBoxVersion: nonEmptyOrNA(buildinfo.SingBoxVersion),
	}}, nil
}

func (s *Server) GetSystemSettings(ctx context.Context, _ *panelv1.GetSystemSettingsRequest) (*panelv1.GetSystemSettingsResponse, error) {
	subscriptionBaseURL, err := s.store.GetSystemSetting(ctx, "subscription_base_url")
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, connectErrorFromHTTP(500, "load settings failed")
	}

	timezone := currentTimezoneName()
	savedTimezone, err := s.store.GetSystemSetting(ctx, "timezone")
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, connectErrorFromHTTP(500, "load settings failed")
	}
	if err == nil {
		if normalized, _, tzErr := normalizeTimezone(savedTimezone); tzErr == nil {
			timezone = normalized
		}
	}

	return &panelv1.GetSystemSettingsResponse{Data: &panelv1.SystemSettings{
		SubscriptionBaseUrl: subscriptionBaseURL,
		Timezone:            timezone,
	}}, nil
}

func (s *Server) UpdateSystemSettings(ctx context.Context, req *panelv1.UpdateSystemSettingsRequest) (*panelv1.UpdateSystemSettingsResponse, error) {
	subURL := strings.TrimSpace(req.GetSubscriptionBaseUrl())
	tz := strings.TrimSpace(req.GetTimezone())

	normalizedURL, err := normalizeSubscriptionBaseURL(subURL)
	if err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}

	normalizedTZ, loc, err := normalizeTimezone(tz)
	if err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}

	if normalizedURL == "" {
		if err := s.store.DeleteSystemSetting(ctx, "subscription_base_url"); err != nil {
			return nil, connectErrorFromHTTP(500, "save settings failed")
		}
	} else {
		if err := s.store.UpsertSystemSetting(ctx, "subscription_base_url", normalizedURL); err != nil {
			return nil, connectErrorFromHTTP(500, "save settings failed")
		}
	}

	if err := s.store.UpsertSystemSetting(ctx, "timezone", normalizedTZ); err != nil {
		return nil, connectErrorFromHTTP(500, "save settings failed")
	}

	time.Local = loc

	return &panelv1.UpdateSystemSettingsResponse{Data: &panelv1.SystemSettings{
		SubscriptionBaseUrl: normalizedURL,
		Timezone:            normalizedTZ,
	}}, nil
}

const (
	rpcDefaultListLimit = 50
	rpcMaxListLimit     = 500

	rpcSyncJobKeepPerNode = 500

	rpcTriggerManualNodeSync = "manual_node_sync"
	rpcTriggerRetry          = "manual_retry"
	rpcTriggerInbound        = "auto_inbound_change"
	rpcTriggerUser           = "auto_user_change"
	rpcTriggerGroup          = "auto_group_change"
)

func rpcPagination(limit *int32, offset *int32) (int, int, error) {
	l := rpcDefaultListLimit
	if limit != nil {
		l = int(*limit)
		if l < 0 || l > rpcMaxListLimit {
			return 0, 0, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid pagination"))
		}
	}
	o := 0
	if offset != nil {
		o = int(*offset)
		if o < 0 {
			return 0, 0, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid pagination"))
		}
	}
	return l, o, nil
}

func listUsersForStatusRPC(ctx context.Context, store *db.Store, status string, limit int, offset int) ([]db.User, error) {
	if status == "" {
		return store.ListUsers(ctx, limit, offset, "")
	}
	if status == userstate.StatusActive || status == userstate.StatusDisabled || status == userstate.StatusExpired || status == userstate.StatusTrafficExceeded {
		return store.ListUsersByEffectiveStatus(ctx, limit, offset, status, time.Now().UTC())
	}
	return nil, errors.New("invalid status")
}

func parseUserUpdateRPC(req *panelv1.UpdateUserRequest) (db.UserUpdate, error) {
	update := db.UserUpdate{}
	if req.Username != nil {
		name := strings.TrimSpace(*req.Username)
		if name == "" {
			return update, errors.New("invalid username")
		}
		update.Username = &name
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != userstate.StatusActive && st != userstate.StatusDisabled && st != userstate.StatusExpired && st != userstate.StatusTrafficExceeded {
			return update, errors.New("invalid status")
		}
		update.Status = &st
	}
	if req.ExpireAt != nil {
		update.ExpireAtSet = true
		raw := strings.TrimSpace(*req.ExpireAt)
		if raw == "" {
			update.ExpireAt = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return update, errors.New("invalid expire_at")
			}
			update.ExpireAt = &parsed
		}
	}
	if req.TrafficLimit != nil {
		if *req.TrafficLimit < 0 {
			return update, errors.New("invalid traffic_limit")
		}
		update.TrafficLimit = req.TrafficLimit
	}
	if req.TrafficResetDay != nil {
		v := int(*req.TrafficResetDay)
		if v < 0 || v > 31 {
			return update, errors.New("invalid traffic_reset_day")
		}
		update.TrafficResetDay = &v
	}
	return update, nil
}

var rpcNodeSyncLocks = struct {
	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}{locks: map[int64]*sync.Mutex{}}

func rpcNodeLock(nodeID int64) *sync.Mutex {
	rpcNodeSyncLocks.mu.Lock()
	defer rpcNodeSyncLocks.mu.Unlock()
	if lock, ok := rpcNodeSyncLocks.locks[nodeID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	rpcNodeSyncLocks.locks[nodeID] = lock
	return lock
}

func parseSyncHTTPStatus(err error) int {
	if err == nil {
		return 200
	}
	msg := strings.TrimSpace(err.Error())
	const prefix = "node sync status "
	if !strings.HasPrefix(msg, prefix) {
		return 0
	}
	remain := strings.TrimPrefix(msg, prefix)
	idx := strings.Index(remain, ":")
	if idx <= 0 {
		return 0
	}
	code, convErr := strconv.Atoi(strings.TrimSpace(remain[:idx]))
	if convErr != nil || code < 100 || code > 599 {
		return 0
	}
	return code
}

func normalizeSyncClientError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(msg, "node sync status ") {
		return msg
	}
	return "node sync request failed: " + msg
}

func (s *Server) runNodeSync(ctx context.Context, n db.Node, triggerSource string, parentJobID *int64) syncResult {
	lock := rpcNodeLock(n.ID)
	lock.Lock()
	defer lock.Unlock()

	if triggerSource == "" {
		triggerSource = rpcTriggerManualNodeSync
	}

	job, err := s.store.CreateSyncJob(ctx, db.SyncJobCreate{NodeID: n.ID, ParentJobID: parentJobID, TriggerSource: triggerSource})
	if err != nil {
		msg := "create sync job failed"
		return syncResult{Status: "error", Error: &msg}
	}
	defer func() {
		_ = s.store.PruneSyncJobsByNode(ctx, n.ID, rpcSyncJobKeepPerNode)
	}()

	if n.GroupID == nil {
		msg := "node group_id not set"
		_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusFailed, AttemptCount: 0, ErrorSummary: msg, FinishedAt: s.store.NowUTC()})
		return syncResult{Status: "error", Error: &msg}
	}

	inbounds, err := s.store.ListInbounds(ctx, 10000, 0, n.ID)
	if err != nil {
		msg := "list inbounds failed"
		_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusFailed, AttemptCount: 0, ErrorSummary: msg, FinishedAt: s.store.NowUTC()})
		return syncResult{Status: "error", Error: &msg}
	}
	users, err := s.store.ListActiveUsersForGroup(ctx, *n.GroupID)
	if err != nil {
		msg := "list users failed"
		_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusFailed, AttemptCount: 0, ErrorSummary: msg, FinishedAt: s.store.NowUTC()})
		return syncResult{Status: "error", Error: &msg}
	}
	payload, err := node.BuildSyncPayload(n, inbounds, users)
	if err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = "build payload failed"
		}
		_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusFailed, AttemptCount: 0, ErrorSummary: msg, FinishedAt: s.store.NowUTC()})
		return syncResult{Status: "error", Error: &msg}
	}

	now := s.store.NowUTC()
	_ = s.store.UpdateSyncJobStart(ctx, job.ID, db.SyncJobStartUpdate{InboundCount: len(inbounds), ActiveUserCount: len(users), StartedAt: now})
	attempt, _ := s.store.CreateSyncAttempt(ctx, db.SyncAttemptCreate{JobID: job.ID, AttemptNo: 1, Status: db.SyncAttemptStatusRunning, StartedAt: now})

	client := node.NewClient(nil)
	err = client.SyncConfig(ctx, n, payload)
	finishedAt := s.store.NowUTC()
	if err == nil {
		if attempt.ID > 0 {
			_ = s.store.UpdateSyncAttemptFinish(ctx, attempt.ID, db.SyncAttemptFinishUpdate{Status: db.SyncAttemptStatusSuccess, HTTPStatus: 200, FinishedAt: finishedAt})
		}
		_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusSuccess, AttemptCount: 1, FinishedAt: finishedAt})
		return syncResult{Status: "ok"}
	}

	errMsg := normalizeSyncClientError(err)
	httpStatus := parseSyncHTTPStatus(err)
	if attempt.ID > 0 {
		_ = s.store.UpdateSyncAttemptFinish(ctx, attempt.ID, db.SyncAttemptFinishUpdate{Status: db.SyncAttemptStatusFailed, HTTPStatus: httpStatus, ErrorSummary: errMsg, FinishedAt: finishedAt})
	}
	_ = s.store.UpdateSyncJobFinish(ctx, job.ID, db.SyncJobFinishUpdate{Status: db.SyncJobStatusFailed, AttemptCount: 1, ErrorSummary: errMsg, FinishedAt: finishedAt})
	return syncResult{Status: "error", Error: &errMsg}
}

func uniquePositiveInt64(items []int64) []int64 {
	out := make([]int64, 0, len(items))
	seen := map[int64]struct{}{}
	for _, id := range items {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *Server) syncNodesByGroupIDs(ctx context.Context, groupIDs []int64, triggerSource string) {
	uniqueGroups := uniquePositiveInt64(groupIDs)
	if len(uniqueGroups) == 0 {
		return
	}

	groupSet := make(map[int64]struct{}, len(uniqueGroups))
	for _, id := range uniqueGroups {
		groupSet[id] = struct{}{}
	}

	offset := 0
	for {
		nodes, err := s.store.ListNodes(ctx, 200, offset)
		if err != nil {
			return
		}
		for _, n := range nodes {
			if n.GroupID == nil {
				continue
			}
			if _, ok := groupSet[*n.GroupID]; !ok {
				continue
			}
			res := s.runNodeSync(ctx, n, triggerSource, nil)
			if res.Status == "ok" {
				_ = s.store.MarkNodeOnline(ctx, n.ID, s.store.NowUTC())
			} else if res.Error != nil && strings.HasPrefix(strings.TrimSpace(*res.Error), "node sync request failed:") {
				_ = s.store.MarkNodeOffline(ctx, n.ID)
			}
		}
		if len(nodes) < 200 {
			return
		}
		offset += len(nodes)
	}
}

func parseWindowOrDefaultRPC(raw string, def time.Duration) time.Duration {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return def
	}
	if s == "all" {
		return 0
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || n <= 0 {
			return -1
		}
		return clampWindowRPC(time.Duration(n) * 24 * time.Hour)
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return -1
	}
	return clampWindowRPC(d)
}

func clampWindowRPC(d time.Duration) time.Duration {
	if d < time.Minute {
		return -1
	}
	if d > 90*24*time.Hour {
		return 90 * 24 * time.Hour
	}
	return d
}

func parseJSONMapString(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, nil
	}
	m := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func mapToRawJSON(m map[string]any) json.RawMessage {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func (s *Server) ListUsers(ctx context.Context, req *panelv1.ListUsersRequest) (*panelv1.ListUsersResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	status := ""
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
	}
	users, err := listUsersForStatusRPC(ctx, s.store, status, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			return nil, connectErrorFromHTTP(400, "invalid status")
		}
		return nil, connectErrorFromHTTP(500, "list users failed")
	}

	ids := make([]int64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	groupMap, err := s.store.ListUserGroupIDsBatch(ctx, ids)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "load user groups failed")
	}

	out := make([]*panelv1.User, 0, len(users))
	for _, user := range users {
		out = append(out, mapDBUser(user, groupMap[user.ID]))
	}
	return &panelv1.ListUsersResponse{Data: out}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *panelv1.CreateUserRequest) (*panelv1.CreateUserResponse, error) {
	username := strings.TrimSpace(req.GetUsername())
	if username == "" {
		return nil, connectErrorFromHTTP(400, "invalid username")
	}

	user, err := s.store.CreateUser(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "username already exists")
		}
		return nil, connectErrorFromHTTP(500, "create user failed")
	}
	groups, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "load user groups failed")
	}
	return &panelv1.CreateUserResponse{Data: mapDBUser(user, groups)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *panelv1.GetUserRequest) (*panelv1.GetUserResponse, error) {
	user, err := s.store.GetUserByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "get user failed")
	}
	groups, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "load user groups failed")
	}
	return &panelv1.GetUserResponse{Data: mapDBUser(user, groups)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *panelv1.UpdateUserRequest) (*panelv1.UpdateUserResponse, error) {
	update, err := parseUserUpdateRPC(req)
	if err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}

	user, err := s.store.UpdateUser(ctx, req.Id, update)
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "username already exists")
		}
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "update user failed")
	}

	groupIDs, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err == nil {
		s.syncNodesByGroupIDs(ctx, groupIDs, rpcTriggerUser)
	}

	groups, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "load user groups failed")
	}
	return &panelv1.UpdateUserResponse{Data: mapDBUser(user, groups)}, nil
}

func (s *Server) DisableUser(ctx context.Context, req *panelv1.DisableUserRequest) (*panelv1.DisableUserResponse, error) {
	if err := s.store.DisableUser(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "disable user failed")
	}

	user, err := s.store.GetUserByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "get user failed")
	}
	groupIDs, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err == nil {
		s.syncNodesByGroupIDs(ctx, groupIDs, rpcTriggerUser)
	}
	groups, err := s.store.ListUserGroupIDs(ctx, user.ID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "load user groups failed")
	}
	return &panelv1.DisableUserResponse{Data: mapDBUser(user, groups)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *panelv1.DeleteUserRequest) (*panelv1.DeleteUserResponse, error) {
	groupIDs, err := s.store.ListUserGroupIDs(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list user groups failed")
	}
	if err := s.store.DeleteUser(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "delete user failed")
	}
	s.syncNodesByGroupIDs(ctx, groupIDs, rpcTriggerUser)
	msg := "user deleted"
	return &panelv1.DeleteUserResponse{Message: &msg}, nil
}

func (s *Server) GetUserGroups(ctx context.Context, req *panelv1.GetUserGroupsRequest) (*panelv1.GetUserGroupsResponse, error) {
	if _, err := s.store.GetUserByID(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "user not found")
		}
		return nil, connectErrorFromHTTP(500, "get user failed")
	}
	ids, err := s.store.ListUserGroupIDs(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list user groups failed")
	}
	return &panelv1.GetUserGroupsResponse{GroupIds: append([]int64{}, ids...)}, nil
}

func (s *Server) ReplaceUserGroups(ctx context.Context, req *panelv1.ReplaceUserGroupsRequest) (*panelv1.ReplaceUserGroupsResponse, error) {
	previousGroupIDs, err := s.store.ListUserGroupIDs(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list user groups failed")
	}
	if err := s.store.ReplaceUserGroups(ctx, req.Id, req.GroupIds); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		return nil, connectErrorFromHTTP(400, "invalid group_ids")
	}
	ids, err := s.store.ListUserGroupIDs(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list user groups failed")
	}
	syncGroupIDs := append([]int64{}, previousGroupIDs...)
	syncGroupIDs = append(syncGroupIDs, ids...)
	s.syncNodesByGroupIDs(ctx, syncGroupIDs, rpcTriggerUser)
	return &panelv1.ReplaceUserGroupsResponse{GroupIds: append([]int64{}, ids...)}, nil
}

func (s *Server) ListGroups(ctx context.Context, req *panelv1.ListGroupsRequest) (*panelv1.ListGroupsResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	groups, err := s.store.ListGroups(ctx, limit, offset)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list groups failed")
	}
	out := make([]*panelv1.Group, 0, len(groups))
	for _, group := range groups {
		out = append(out, mapDBGroup(group))
	}
	return &panelv1.ListGroupsResponse{Data: out}, nil
}

func (s *Server) CreateGroup(ctx context.Context, req *panelv1.CreateGroupRequest) (*panelv1.CreateGroupResponse, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, connectErrorFromHTTP(400, "invalid name")
	}
	desc := ""
	if req.Description != nil {
		desc = strings.TrimSpace(*req.Description)
	}
	group, err := s.store.CreateGroup(ctx, name, desc)
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "group name already exists")
		}
		return nil, connectErrorFromHTTP(500, "create group failed")
	}
	return &panelv1.CreateGroupResponse{Data: mapDBGroup(group)}, nil
}

func (s *Server) GetGroup(ctx context.Context, req *panelv1.GetGroupRequest) (*panelv1.GetGroupResponse, error) {
	group, err := s.store.GetGroupByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "group not found")
		}
		return nil, connectErrorFromHTTP(500, "get group failed")
	}
	return &panelv1.GetGroupResponse{Data: mapDBGroup(group)}, nil
}

func (s *Server) UpdateGroup(ctx context.Context, req *panelv1.UpdateGroupRequest) (*panelv1.UpdateGroupResponse, error) {
	update := db.GroupUpdate{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, connectErrorFromHTTP(400, "invalid name")
		}
		update.Name = &name
	}
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		update.Description = &desc
	}
	group, err := s.store.UpdateGroup(ctx, req.Id, update)
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "group name already exists")
		}
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "group not found")
		}
		return nil, connectErrorFromHTTP(500, "update group failed")
	}
	return &panelv1.UpdateGroupResponse{Data: mapDBGroup(group)}, nil
}

func (s *Server) DeleteGroup(ctx context.Context, req *panelv1.DeleteGroupRequest) (*panelv1.DeleteGroupResponse, error) {
	if err := s.store.DeleteGroup(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "group not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "group is in use")
		}
		return nil, connectErrorFromHTTP(500, "delete group failed")
	}
	return &panelv1.DeleteGroupResponse{Status: "ok"}, nil
}

func (s *Server) ListGroupUsers(ctx context.Context, req *panelv1.ListGroupUsersRequest) (*panelv1.ListGroupUsersResponse, error) {
	users, err := s.store.ListGroupUsers(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "group not found")
		}
		return nil, connectErrorFromHTTP(500, "list group users failed")
	}
	out := make([]*panelv1.GroupUsersListItem, 0, len(users))
	for _, user := range users {
		out = append(out, &panelv1.GroupUsersListItem{
			Id:           user.ID,
			Uuid:         user.UUID,
			Username:     user.Username,
			TrafficLimit: user.TrafficLimit,
			TrafficUsed:  user.TrafficUsed,
			Status:       userstate.EffectiveStatus(user, time.Now().UTC()),
		})
	}
	return &panelv1.ListGroupUsersResponse{Data: out}, nil
}

func (s *Server) ReplaceGroupUsers(ctx context.Context, req *panelv1.ReplaceGroupUsersRequest) (*panelv1.ReplaceGroupUsersResponse, error) {
	if err := s.store.ReplaceGroupUsers(ctx, req.Id, req.UserIds); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		return nil, connectErrorFromHTTP(400, "invalid user_ids")
	}
	s.syncNodesByGroupIDs(ctx, []int64{req.Id}, rpcTriggerGroup)
	return &panelv1.ReplaceGroupUsersResponse{UserIds: uniquePositiveInt64(req.UserIds)}, nil
}

func (s *Server) ListNodes(ctx context.Context, req *panelv1.ListNodesRequest) (*panelv1.ListNodesResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	nodes, err := s.store.ListNodes(ctx, limit, offset)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list nodes failed")
	}
	out := make([]*panelv1.Node, 0, len(nodes))
	for _, item := range nodes {
		out = append(out, mapDBNode(item))
	}
	return &panelv1.ListNodesResponse{Data: out}, nil
}

func (s *Server) CreateNode(ctx context.Context, req *panelv1.CreateNodeRequest) (*panelv1.CreateNodeResponse, error) {
	name := strings.TrimSpace(req.GetName())
	apiAddress := strings.TrimSpace(req.GetApiAddress())
	publicAddress := strings.TrimSpace(req.GetPublicAddress())
	if name == "" || apiAddress == "" || req.GetApiPort() <= 0 || strings.TrimSpace(req.GetSecretKey()) == "" || publicAddress == "" {
		return nil, connectErrorFromHTTP(400, "invalid node")
	}
	nodeItem, err := s.store.CreateNode(ctx, db.NodeCreate{
		Name:          name,
		APIAddress:    apiAddress,
		APIPort:       int(req.GetApiPort()),
		SecretKey:     req.GetSecretKey(),
		PublicAddress: publicAddress,
		GroupID:       req.GroupId,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		return nil, connectErrorFromHTTP(500, "create node failed")
	}
	return &panelv1.CreateNodeResponse{Data: mapDBNode(nodeItem)}, nil
}

func (s *Server) GetNode(ctx context.Context, req *panelv1.GetNodeRequest) (*panelv1.GetNodeResponse, error) {
	nodeItem, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "get node failed")
	}
	return &panelv1.GetNodeResponse{Data: mapDBNode(nodeItem)}, nil
}

func (s *Server) UpdateNode(ctx context.Context, req *panelv1.UpdateNodeRequest) (*panelv1.UpdateNodeResponse, error) {
	update := db.NodeUpdate{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, connectErrorFromHTTP(400, "invalid name")
		}
		update.Name = &name
	}
	if req.ApiAddress != nil {
		addr := strings.TrimSpace(*req.ApiAddress)
		if addr == "" {
			return nil, connectErrorFromHTTP(400, "invalid api_address")
		}
		update.APIAddress = &addr
	}
	if req.ApiPort != nil {
		if *req.ApiPort <= 0 {
			return nil, connectErrorFromHTTP(400, "invalid api_port")
		}
		v := int(*req.ApiPort)
		update.APIPort = &v
	}
	if req.SecretKey != nil {
		if strings.TrimSpace(*req.SecretKey) == "" {
			return nil, connectErrorFromHTTP(400, "invalid secret_key")
		}
		update.SecretKey = req.SecretKey
	}
	if req.PublicAddress != nil {
		addr := strings.TrimSpace(*req.PublicAddress)
		if addr == "" {
			return nil, connectErrorFromHTTP(400, "invalid public_address")
		}
		update.PublicAddress = &addr
	}
	if req.ClearGroupId {
		update.GroupIDSet = true
		update.GroupID = nil
	} else if req.GroupId != nil {
		if *req.GroupId <= 0 {
			return nil, connectErrorFromHTTP(400, "invalid group_id")
		}
		update.GroupIDSet = true
		update.GroupID = req.GroupId
	}

	nodeItem, err := s.store.UpdateNode(ctx, req.Id, update)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		return nil, connectErrorFromHTTP(500, "update node failed")
	}
	return &panelv1.UpdateNodeResponse{Data: mapDBNode(nodeItem)}, nil
}

func (s *Server) DeleteNode(ctx context.Context, req *panelv1.DeleteNodeRequest) (*panelv1.DeleteNodeResponse, error) {
	if !req.Force {
		if err := s.store.DeleteNode(ctx, req.Id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return nil, connectErrorFromHTTP(404, "node not found")
			}
			if errors.Is(err, db.ErrConflict) {
				return nil, connectErrorFromHTTP(409, "node is in use")
			}
			return nil, connectErrorFromHTTP(500, "delete node failed")
		}
		return &panelv1.DeleteNodeResponse{Status: "ok"}, nil
	}

	nodeItem, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "get node failed")
	}

	inbounds, err := s.store.ListInbounds(ctx, 10000, 0, nodeItem.ID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list inbounds failed")
	}
	if len(inbounds) > 0 {
		lock := rpcNodeLock(nodeItem.ID)
		lock.Lock()
		emptyPayload := node.SyncPayload{Inbounds: []map[string]any{}}
		syncErr := node.NewClient(nil).SyncConfig(ctx, nodeItem, emptyPayload)
		lock.Unlock()
		if syncErr != nil {
			return nil, connectErrorFromHTTP(502, "force drain failed: "+syncErr.Error())
		}
		_ = s.store.MarkNodeOnline(ctx, nodeItem.ID, s.store.NowUTC())
	}

	deletedInbounds, err := s.store.DeleteInboundsByNode(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "delete node inbounds failed")
	}
	if err := s.store.DeleteNode(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "node is in use")
		}
		return nil, connectErrorFromHTTP(500, "delete node failed")
	}
	forceTrue := true
	del := int32(deletedInbounds)
	return &panelv1.DeleteNodeResponse{Status: "ok", Force: &forceTrue, DeletedInbounds: &del}, nil
}

func (s *Server) GetNodeHealth(ctx context.Context, req *panelv1.GetNodeHealthRequest) (*panelv1.GetNodeHealthResponse, error) {
	nodeItem, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "get node failed")
	}
	if err := node.NewClient(nil).Health(ctx, nodeItem); err != nil {
		_ = s.store.MarkNodeOffline(ctx, nodeItem.ID)
		return nil, connectErrorFromHTTP(502, err.Error())
	}
	_ = s.store.MarkNodeOnline(ctx, nodeItem.ID, s.store.NowUTC())
	return &panelv1.GetNodeHealthResponse{Status: "ok"}, nil
}

func (s *Server) SyncNode(ctx context.Context, req *panelv1.SyncNodeRequest) (*panelv1.SyncNodeResponse, error) {
	nodeItem, err := s.store.GetNodeByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "get node failed")
	}
	if nodeItem.GroupID == nil {
		return nil, connectErrorFromHTTP(400, "node group_id not set")
	}
	res := s.runNodeSync(ctx, nodeItem, rpcTriggerManualNodeSync, nil)
	if res.Status != "ok" {
		if res.Error == nil {
			return nil, connectErrorFromHTTP(502, "sync failed")
		}
		return nil, connectErrorFromHTTP(502, *res.Error)
	}
	_ = s.store.MarkNodeOnline(ctx, nodeItem.ID, s.store.NowUTC())
	return &panelv1.SyncNodeResponse{Status: "ok"}, nil
}

func (s *Server) ListNodeTraffic(ctx context.Context, req *panelv1.ListNodeTrafficRequest) (*panelv1.ListNodeTrafficResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListNodeTrafficSamples(ctx, req.Id, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "invalid node_id") {
			return nil, connectErrorFromHTTP(400, "invalid id")
		}
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "not found")
		}
		return nil, connectErrorFromHTTP(500, err.Error())
	}
	out := make([]*panelv1.NodeTrafficSample, 0, len(items))
	for _, item := range items {
		out = append(out, &panelv1.NodeTrafficSample{
			Id:         item.ID,
			InboundTag: item.InboundTag,
			Upload:     item.Upload,
			Download:   item.Download,
			RecordedAt: formatTime(item.RecordedAt),
		})
	}
	return &panelv1.ListNodeTrafficResponse{Data: out}, nil
}

func (s *Server) GetTrafficNodesSummary(ctx context.Context, req *panelv1.GetTrafficNodesSummaryRequest) (*panelv1.GetTrafficNodesSummaryResponse, error) {
	window := parseWindowOrDefaultRPC(req.GetWindow(), 24*time.Hour)
	if window < 0 {
		return nil, connectErrorFromHTTP(400, "invalid window")
	}
	p := traffic.NewSQLiteProvider(s.store)
	items, err := p.NodesSummary(ctx, window)
	if err != nil {
		return nil, connectErrorFromHTTP(500, err.Error())
	}
	out := make([]*panelv1.TrafficNodeSummary, 0, len(items))
	for _, item := range items {
		out = append(out, &panelv1.TrafficNodeSummary{
			NodeId:         item.NodeID,
			Upload:         item.Upload,
			Download:       item.Download,
			LastRecordedAt: formatTime(item.LastRecordedAt),
			Samples:        item.Samples,
			Inbounds:       item.Inbounds,
		})
	}
	return &panelv1.GetTrafficNodesSummaryResponse{Data: out}, nil
}

func (s *Server) GetTrafficTotalSummary(ctx context.Context, req *panelv1.GetTrafficTotalSummaryRequest) (*panelv1.GetTrafficTotalSummaryResponse, error) {
	window := parseWindowOrDefaultRPC(req.GetWindow(), 24*time.Hour)
	if window < 0 {
		return nil, connectErrorFromHTTP(400, "invalid window")
	}
	p := traffic.NewSQLiteProvider(s.store)
	item, err := p.TotalSummary(ctx, window)
	if err != nil {
		return nil, connectErrorFromHTTP(500, err.Error())
	}
	return &panelv1.GetTrafficTotalSummaryResponse{Data: &panelv1.TrafficTotalSummary{
		Upload:         item.Upload,
		Download:       item.Download,
		LastRecordedAt: formatTime(item.LastRecordedAt),
		Samples:        item.Samples,
		Nodes:          item.Nodes,
		Inbounds:       item.Inbounds,
	}}, nil
}

func (s *Server) GetTrafficTimeseries(ctx context.Context, req *panelv1.GetTrafficTimeseriesRequest) (*panelv1.GetTrafficTimeseriesResponse, error) {
	window := parseWindowOrDefaultRPC(req.GetWindow(), 24*time.Hour)
	if window < 0 {
		return nil, connectErrorFromHTTP(400, "invalid window")
	}
	bucket := traffic.BucketHour
	if req.Bucket != nil {
		switch strings.TrimSpace(strings.ToLower(*req.Bucket)) {
		case string(traffic.BucketMinute):
			bucket = traffic.BucketMinute
		case string(traffic.BucketHour):
			bucket = traffic.BucketHour
		case string(traffic.BucketDay):
			bucket = traffic.BucketDay
		default:
			return nil, connectErrorFromHTTP(400, "invalid bucket")
		}
	}
	nodeID := int64(0)
	if req.NodeId != nil {
		if *req.NodeId < 0 {
			return nil, connectErrorFromHTTP(400, "invalid node_id")
		}
		nodeID = *req.NodeId
	}
	p := traffic.NewSQLiteProvider(s.store)
	items, err := p.Timeseries(ctx, traffic.TimeseriesQuery{Window: window, Bucket: bucket, NodeID: nodeID})
	if err != nil {
		return nil, connectErrorFromHTTP(500, err.Error())
	}
	out := make([]*panelv1.TrafficTimeseriesPoint, 0, len(items))
	for _, item := range items {
		out = append(out, &panelv1.TrafficTimeseriesPoint{
			BucketStart: formatTime(item.BucketStart),
			Upload:      item.Upload,
			Download:    item.Download,
		})
	}
	return &panelv1.GetTrafficTimeseriesResponse{Data: out}, nil
}

func (s *Server) ListInbounds(ctx context.Context, req *panelv1.ListInboundsRequest) (*panelv1.ListInboundsResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	nodeID := int64(0)
	if req.NodeId != nil {
		nodeID = *req.NodeId
		if nodeID < 0 {
			return nil, connectErrorFromHTTP(400, "invalid node_id")
		}
	}
	inbounds, err := s.store.ListInbounds(ctx, limit, offset, nodeID)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list inbounds failed")
	}
	out := make([]*panelv1.Inbound, 0, len(inbounds))
	for _, inb := range inbounds {
		out = append(out, mapDBInbound(inb))
	}
	return &panelv1.ListInboundsResponse{Data: out}, nil
}

func (s *Server) CreateInbound(ctx context.Context, req *panelv1.CreateInboundRequest) (*panelv1.CreateInboundResponse, error) {
	tag := strings.TrimSpace(req.GetTag())
	protocol := strings.TrimSpace(req.GetProtocol())
	if tag == "" || protocol == "" || req.GetNodeId() <= 0 || req.GetListenPort() <= 0 {
		return nil, connectErrorFromHTTP(400, "invalid inbound")
	}

	settingsMap, err := parseJSONMapString(req.GetSettingsJson())
	if err != nil || len(settingsMap) == 0 {
		return nil, connectErrorFromHTTP(400, "invalid settings")
	}
	if err := inbval.ValidateSettings(protocol, settingsMap); err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}

	tlsMap := map[string]any{}
	if req.TlsSettingsJson != nil {
		tlsMap, err = parseJSONMapString(req.GetTlsSettingsJson())
		if err != nil {
			return nil, connectErrorFromHTTP(400, "invalid tls_settings")
		}
	}
	transportMap := map[string]any{}
	if req.TransportSettingsJson != nil {
		transportMap, err = parseJSONMapString(req.GetTransportSettingsJson())
		if err != nil {
			return nil, connectErrorFromHTTP(400, "invalid transport_settings")
		}
	}

	publicPort := int(req.GetListenPort())
	if req.PublicPort != nil {
		publicPort = int(*req.PublicPort)
	}

	inb, err := s.store.CreateInbound(ctx, db.InboundCreate{
		Tag:               tag,
		NodeID:            req.GetNodeId(),
		Protocol:          protocol,
		ListenPort:        int(req.GetListenPort()),
		PublicPort:        publicPort,
		Settings:          mapToRawJSON(settingsMap),
		TLSSettings:       mapToRawJSON(tlsMap),
		TransportSettings: mapToRawJSON(transportMap),
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "create inbound failed")
	}

	nodeItem, err := s.store.GetNodeByID(ctx, inb.NodeID)
	if err != nil {
		msg := "get node failed"
		return &panelv1.CreateInboundResponse{Data: mapDBInbound(inb), Sync: mapSyncResult(syncResult{Status: "error", Error: &msg})}, nil
	}
	res := s.runNodeSync(ctx, nodeItem, rpcTriggerInbound, nil)
	return &panelv1.CreateInboundResponse{Data: mapDBInbound(inb), Sync: mapSyncResult(res)}, nil
}

func (s *Server) GetInbound(ctx context.Context, req *panelv1.GetInboundRequest) (*panelv1.GetInboundResponse, error) {
	inb, err := s.store.GetInboundByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "inbound not found")
		}
		return nil, connectErrorFromHTTP(500, "get inbound failed")
	}
	return &panelv1.GetInboundResponse{Data: mapDBInbound(inb)}, nil
}

func (s *Server) UpdateInbound(ctx context.Context, req *panelv1.UpdateInboundRequest) (*panelv1.UpdateInboundResponse, error) {
	cur, err := s.store.GetInboundByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "inbound not found")
		}
		return nil, connectErrorFromHTTP(500, "get inbound failed")
	}

	upd := db.InboundUpdate{}
	if req.Tag != nil {
		tag := strings.TrimSpace(*req.Tag)
		if tag == "" {
			return nil, connectErrorFromHTTP(400, "invalid tag")
		}
		upd.Tag = &tag
	}
	if req.Protocol != nil {
		p := strings.TrimSpace(*req.Protocol)
		if p == "" {
			return nil, connectErrorFromHTTP(400, "invalid protocol")
		}
		upd.Protocol = &p
	}
	if req.ListenPort != nil {
		if *req.ListenPort <= 0 {
			return nil, connectErrorFromHTTP(400, "invalid listen_port")
		}
		v := int(*req.ListenPort)
		upd.ListenPort = &v
	}
	if req.PublicPort != nil {
		if *req.PublicPort < 0 {
			return nil, connectErrorFromHTTP(400, "invalid public_port")
		}
		v := int(*req.PublicPort)
		upd.PublicPort = &v
	}
	if req.SettingsJson != nil {
		settingsMap, parseErr := parseJSONMapString(req.GetSettingsJson())
		if parseErr != nil || len(settingsMap) == 0 {
			return nil, connectErrorFromHTTP(400, "invalid settings")
		}
		raw := mapToRawJSON(settingsMap)
		upd.Settings = &raw
	}
	if req.TlsSettingsJson != nil {
		m, parseErr := parseJSONMapString(req.GetTlsSettingsJson())
		if parseErr != nil {
			return nil, connectErrorFromHTTP(400, "invalid tls_settings")
		}
		raw := mapToRawJSON(m)
		upd.TLSSettings = &raw
	}
	if req.TransportSettingsJson != nil {
		m, parseErr := parseJSONMapString(req.GetTransportSettingsJson())
		if parseErr != nil {
			return nil, connectErrorFromHTTP(400, "invalid transport_settings")
		}
		raw := mapToRawJSON(m)
		upd.TransportSettings = &raw
	}

	finalProtocol := cur.Protocol
	if upd.Protocol != nil {
		finalProtocol = *upd.Protocol
	}
	finalSettings := cur.Settings
	if upd.Settings != nil {
		finalSettings = *upd.Settings
	}
	if len(finalSettings) == 0 || !json.Valid(finalSettings) {
		return nil, connectErrorFromHTTP(400, "invalid settings")
	}
	settingsMap := map[string]any{}
	if err := json.Unmarshal(finalSettings, &settingsMap); err != nil {
		return nil, connectErrorFromHTTP(400, "invalid settings")
	}
	if err := inbval.ValidateSettings(finalProtocol, settingsMap); err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}

	inb, err := s.store.UpdateInbound(ctx, req.Id, upd)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "inbound not found")
		}
		if errors.Is(err, db.ErrConflict) {
			return nil, connectErrorFromHTTP(409, "conflict")
		}
		return nil, connectErrorFromHTTP(500, "update inbound failed")
	}
	nodeItem, err := s.store.GetNodeByID(ctx, inb.NodeID)
	if err != nil {
		msg := "get node failed"
		return &panelv1.UpdateInboundResponse{Data: mapDBInbound(inb), Sync: mapSyncResult(syncResult{Status: "error", Error: &msg})}, nil
	}
	res := s.runNodeSync(ctx, nodeItem, rpcTriggerInbound, nil)
	return &panelv1.UpdateInboundResponse{Data: mapDBInbound(inb), Sync: mapSyncResult(res)}, nil
}

func (s *Server) DeleteInbound(ctx context.Context, req *panelv1.DeleteInboundRequest) (*panelv1.DeleteInboundResponse, error) {
	cur, err := s.store.GetInboundByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "inbound not found")
		}
		return nil, connectErrorFromHTTP(500, "get inbound failed")
	}
	if err := s.store.DeleteInbound(ctx, req.Id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "inbound not found")
		}
		return nil, connectErrorFromHTTP(500, "delete inbound failed")
	}
	nodeItem, err := s.store.GetNodeByID(ctx, cur.NodeID)
	if err != nil {
		msg := "get node failed"
		return &panelv1.DeleteInboundResponse{Status: "ok", Sync: mapSyncResult(syncResult{Status: "error", Error: &msg})}, nil
	}
	res := s.runNodeSync(ctx, nodeItem, rpcTriggerInbound, nil)
	return &panelv1.DeleteInboundResponse{Status: "ok", Sync: mapSyncResult(res)}, nil
}

func (s *Server) ListSyncJobs(ctx context.Context, req *panelv1.ListSyncJobsRequest) (*panelv1.ListSyncJobsResponse, error) {
	limit, offset, err := rpcPagination(req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	from, err := parseTimeRFC3339Ptr(req.From)
	if err != nil {
		return nil, err
	}
	to, err := parseTimeRFC3339Ptr(req.To)
	if err != nil {
		return nil, err
	}
	filter := db.SyncJobsListFilter{Limit: limit, Offset: offset, TriggerSource: req.GetTriggerSource(), From: from, To: to}
	if req.NodeId != nil {
		filter.NodeID = *req.NodeId
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != db.SyncJobStatusQueued && st != db.SyncJobStatusRunning && st != db.SyncJobStatusSuccess && st != db.SyncJobStatusFailed {
			return nil, connectErrorFromHTTP(400, "invalid status")
		}
		filter.Status = st
	}
	items, err := s.store.ListSyncJobs(ctx, filter)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list sync jobs failed")
	}
	out := make([]*panelv1.SyncJobListItem, 0, len(items))
	for _, item := range items {
		out = append(out, mapDBSyncJob(item))
	}
	return &panelv1.ListSyncJobsResponse{Data: out}, nil
}

func (s *Server) GetSyncJob(ctx context.Context, req *panelv1.GetSyncJobRequest) (*panelv1.GetSyncJobResponse, error) {
	job, err := s.store.GetSyncJobByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "sync job not found")
		}
		return nil, connectErrorFromHTTP(500, "get sync job failed")
	}
	attemptsRaw, err := s.store.ListSyncAttemptsByJobID(ctx, req.Id)
	if err != nil {
		return nil, connectErrorFromHTTP(500, "list sync attempts failed")
	}
	attempts := make([]*panelv1.SyncAttempt, 0, len(attemptsRaw))
	for _, item := range attemptsRaw {
		attempts = append(attempts, mapDBSyncAttempt(item))
	}
	return &panelv1.GetSyncJobResponse{Data: &panelv1.SyncJobDetail{Job: mapDBSyncJob(job), Attempts: attempts}}, nil
}

func (s *Server) RetrySyncJob(ctx context.Context, req *panelv1.RetrySyncJobRequest) (*panelv1.RetrySyncJobResponse, error) {
	if req.Id <= 0 {
		return nil, connectErrorFromHTTP(400, "invalid id")
	}
	parentJob, err := s.store.GetSyncJobByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "sync job not found")
		}
		return nil, connectErrorFromHTTP(500, "get sync job failed")
	}
	nodeItem, err := s.store.GetNodeByID(ctx, parentJob.NodeID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, connectErrorFromHTTP(404, "node not found")
		}
		return nil, connectErrorFromHTTP(500, "get node failed")
	}
	res := s.runNodeSync(ctx, nodeItem, rpcTriggerRetry, &parentJob.ID)
	if res.Status != "ok" {
		if res.Error == nil {
			return nil, connectErrorFromHTTP(502, "sync failed")
		}
		return nil, connectErrorFromHTTP(502, *res.Error)
	}

	jobs, err := s.store.ListSyncJobs(ctx, db.SyncJobsListFilter{Limit: 1, Offset: 0, NodeID: parentJob.NodeID})
	if err != nil || len(jobs) == 0 {
		status := "ok"
		return &panelv1.RetrySyncJobResponse{Status: &status}, nil
	}
	return &panelv1.RetrySyncJobResponse{Data: mapDBSyncJob(jobs[0])}, nil
}

func (s *Server) FormatSingBox(ctx context.Context, req *panelv1.FormatSingBoxRequest) (*panelv1.FormatSingBoxResponse, error) {
	if req.Data == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing data"))
	}
	mode := ""
	if req.Data.Mode != nil {
		mode = *req.Data.Mode
	}
	wrapped, err := wrapConfigIfNeeded(req.Data.Config, mode)
	if err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}
	formatted, err := singboxcli.New().Format(ctx, wrapped)
	if err != nil {
		return nil, connectErrorFromHTTP(400, strings.TrimSpace(err.Error()))
	}
	return &panelv1.FormatSingBoxResponse{Data: &panelv1.FormatSingBoxResult{Formatted: formatted}}, nil
}

func (s *Server) CheckSingBox(ctx context.Context, req *panelv1.CheckSingBoxRequest) (*panelv1.CheckSingBoxResponse, error) {
	if req.Data == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing data"))
	}
	mode := ""
	if req.Data.Mode != nil {
		mode = *req.Data.Mode
	}
	wrapped, err := wrapConfigIfNeeded(req.Data.Config, mode)
	if err != nil {
		return nil, connectErrorFromHTTP(400, err.Error())
	}
	output, err := singboxcli.New().Check(ctx, wrapped)
	if err != nil {
		return &panelv1.CheckSingBoxResponse{Data: &panelv1.CheckSingBoxResult{Ok: false, Output: strings.TrimSpace(err.Error())}}, nil
	}
	return &panelv1.CheckSingBoxResponse{Data: &panelv1.CheckSingBoxResult{Ok: true, Output: output}}, nil
}

func (s *Server) GenerateSingBox(ctx context.Context, req *panelv1.GenerateSingBoxRequest) (*panelv1.GenerateSingBoxResponse, error) {
	output, err := singboxcli.New().Generate(ctx, req.GetCommand())
	if err != nil {
		if errors.Is(err, singboxcli.ErrInvalidGenerateKind) {
			return nil, connectErrorFromHTTP(400, "invalid generate command")
		}
		return nil, connectErrorFromHTTP(400, strings.TrimSpace(err.Error()))
	}
	return &panelv1.GenerateSingBoxResponse{Data: &panelv1.GenerateSingBoxResult{Output: output}}, nil
}

func hasAnyKey(obj map[string]json.RawMessage, keys ...string) bool {
	for _, key := range keys {
		if _, ok := obj[key]; ok {
			return true
		}
	}
	return false
}

func wrapInboundAsConfig(inboundText string) (string, error) {
	var inbound map[string]any
	if err := json.Unmarshal([]byte(inboundText), &inbound); err != nil {
		return "", errors.New("invalid inbound json")
	}
	delete(inbound, "public_port")
	wrapped := map[string]any{
		"log":       map[string]any{"level": "error"},
		"inbounds":  []any{inbound},
		"outbounds": []map[string]any{{"type": "direct", "tag": "direct"}},
	}
	raw, err := json.Marshal(wrapped)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func normalizeInboundEditorConfig(configText string) (string, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(configText), &config); err != nil {
		return "", errors.New("invalid config json")
	}
	rawInbounds, ok := config["inbounds"]
	if !ok {
		return configText, nil
	}
	inbounds, ok := rawInbounds.([]any)
	if !ok {
		return "", errors.New("inbounds must be an array")
	}
	for idx, item := range inbounds {
		inbound, ok := item.(map[string]any)
		if !ok {
			return "", errors.New("inbounds[" + strconv.Itoa(idx) + "] must be an object")
		}
		delete(inbound, "public_port")
	}
	out, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func wrapConfigIfNeeded(configText string, mode string) (string, error) {
	trimmed := strings.TrimSpace(configText)
	if trimmed == "" {
		return "", errors.New("config required")
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		return "", errors.New("invalid json")
	}
	normalizedMode := strings.ToLower(strings.TrimSpace(mode))
	isFullConfig := hasAnyKey(obj, "inbounds", "outbounds", "route", "dns", "log")

	if normalizedMode == "inbound" {
		if isFullConfig {
			return normalizeInboundEditorConfig(trimmed)
		}
		return wrapInboundAsConfig(trimmed)
	}
	if normalizedMode == "config" {
		return trimmed, nil
	}
	if normalizedMode == "" || normalizedMode == "auto" {
		if isFullConfig {
			return normalizeInboundEditorConfig(trimmed)
		}
		return wrapInboundAsConfig(trimmed)
	}
	return trimmed, nil
}
