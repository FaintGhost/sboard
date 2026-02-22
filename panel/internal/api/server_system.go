package api

import (
  "context"
  "encoding/json"
  "errors"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "sboard/panel/internal/buildinfo"
  "sboard/panel/internal/db"
  "sboard/panel/internal/password"
  "sboard/panel/internal/singboxcli"
  "sboard/panel/internal/subscription"
  "sboard/panel/internal/traffic"
  "sboard/panel/internal/userstate"
)

// --- Health ---

func (s *Server) GetHealth(ctx context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
  return GetHealth200JSONResponse{Status: "ok"}, nil
}

// --- Auth ---

func (s *Server) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
  n, err := db.AdminCount(s.store)
  if err != nil {
    return Login500JSONResponse{errInternal("db error")}, nil
  }
  if n == 0 {
    return Login428JSONResponse{Error: "needs setup"}, nil
  }

  a, ok, err := db.AdminGetByUsername(s.store, request.Body.Username)
  if err != nil {
    return Login500JSONResponse{errInternal("db error")}, nil
  }
  if !ok || !password.Verify(a.PasswordHash, request.Body.Password) {
    return Login401JSONResponse{errUnauthorized("unauthorized")}, nil
  }

  token, exp, err := signAdminToken(s.cfg.JWTSecret)
  if err != nil {
    return Login500JSONResponse{errInternal("sign token failed")}, nil
  }
  return Login200JSONResponse{Data: LoginResponse{
    Token:     token,
    ExpiresAt: formatTimeRFC3339OrEmpty(exp),
  }}, nil
}

// --- Bootstrap ---

func (s *Server) GetBootstrapStatus(ctx context.Context, _ GetBootstrapStatusRequestObject) (GetBootstrapStatusResponseObject, error) {
  n, err := db.AdminCount(s.store)
  if err != nil {
    return GetBootstrapStatus500JSONResponse{errInternal("db error")}, nil
  }
  return GetBootstrapStatus200JSONResponse{Data: BootstrapStatus{NeedsSetup: n == 0}}, nil
}

func (s *Server) Bootstrap(ctx context.Context, request BootstrapRequestObject) (BootstrapResponseObject, error) {
  setupToken := ""
  if request.Body.SetupToken != nil {
    setupToken = strings.TrimSpace(*request.Body.SetupToken)
  }
  if setupToken == "" && request.Params.XSetupToken != nil {
    setupToken = strings.TrimSpace(*request.Params.XSetupToken)
  }

  if strings.TrimSpace(s.cfg.SetupToken) == "" || setupToken != strings.TrimSpace(s.cfg.SetupToken) {
    return Bootstrap401JSONResponse{errUnauthorized("unauthorized")}, nil
  }

  username := strings.TrimSpace(request.Body.Username)
  if username == "" {
    return Bootstrap400JSONResponse{errBadRequest("missing username")}, nil
  }
  if strings.TrimSpace(request.Body.Password) == "" {
    return Bootstrap400JSONResponse{errBadRequest("missing password")}, nil
  }
  if request.Body.Password != request.Body.ConfirmPassword {
    return Bootstrap400JSONResponse{errBadRequest("passwords do not match")}, nil
  }
  if len(request.Body.Password) < 8 {
    return Bootstrap400JSONResponse{errBadRequest("password too short")}, nil
  }

  h, err := password.Hash(request.Body.Password)
  if err != nil {
    return Bootstrap400JSONResponse{errBadRequest("invalid password")}, nil
  }
  created, err := db.AdminCreateIfNone(s.store, username, h)
  if err != nil {
    return Bootstrap500JSONResponse{errInternal("db error")}, nil
  }
  if !created {
    return Bootstrap409JSONResponse{errConflict("already initialized")}, nil
  }
  return Bootstrap200JSONResponse{Data: BootstrapResponse{Ok: true}}, nil
}

// --- Admin Profile ---

func (s *Server) GetAdminProfile(ctx context.Context, _ GetAdminProfileRequestObject) (GetAdminProfileResponseObject, error) {
  admin, ok, err := db.AdminGetFirst(s.store)
  if err != nil {
    return GetAdminProfile500JSONResponse{errInternal("db error")}, nil
  }
  if !ok {
    return GetAdminProfile428JSONResponse{Error: "needs setup"}, nil
  }
  return GetAdminProfile200JSONResponse{Data: AdminProfile{Username: admin.Username}}, nil
}

func (s *Server) UpdateAdminProfile(ctx context.Context, request UpdateAdminProfileRequestObject) (UpdateAdminProfileResponseObject, error) {
  admin, ok, err := db.AdminGetFirst(s.store)
  if err != nil {
    return UpdateAdminProfile500JSONResponse{errInternal("db error")}, nil
  }
  if !ok {
    return UpdateAdminProfile428JSONResponse{Error: "needs setup"}, nil
  }

  if strings.TrimSpace(request.Body.OldPassword) == "" {
    return UpdateAdminProfile400JSONResponse{errBadRequest("missing old_password")}, nil
  }
  if !password.Verify(admin.PasswordHash, request.Body.OldPassword) {
    return UpdateAdminProfile401JSONResponse{errUnauthorized("old password mismatch")}, nil
  }

  newUsername := strings.TrimSpace("")
  if request.Body.NewUsername != nil {
    newUsername = strings.TrimSpace(*request.Body.NewUsername)
  }
  if newUsername == "" {
    newUsername = admin.Username
  }

  newPasswordHash := admin.PasswordHash
  hasPasswordChange := (request.Body.NewPassword != nil && strings.TrimSpace(*request.Body.NewPassword) != "") ||
    (request.Body.ConfirmPassword != nil && strings.TrimSpace(*request.Body.ConfirmPassword) != "")
  if hasPasswordChange {
    newPw := ""
    confirmPw := ""
    if request.Body.NewPassword != nil {
      newPw = *request.Body.NewPassword
    }
    if request.Body.ConfirmPassword != nil {
      confirmPw = *request.Body.ConfirmPassword
    }
    if newPw != confirmPw {
      return UpdateAdminProfile400JSONResponse{errBadRequest("passwords do not match")}, nil
    }
    if len(newPw) < 8 {
      return UpdateAdminProfile400JSONResponse{errBadRequest("password too short")}, nil
    }
    hashed, err := password.Hash(newPw)
    if err != nil {
      return UpdateAdminProfile400JSONResponse{errBadRequest("invalid password")}, nil
    }
    newPasswordHash = hashed
  }

  if newUsername == admin.Username && newPasswordHash == admin.PasswordHash {
    return UpdateAdminProfile400JSONResponse{errBadRequest("no changes")}, nil
  }

  if err := db.AdminUpdateCredentials(s.store, admin.ID, newUsername, newPasswordHash); err != nil {
    if errors.Is(err, db.ErrConflict) {
      return UpdateAdminProfile409JSONResponse{errConflict("username exists")}, nil
    }
    if errors.Is(err, db.ErrNotFound) {
      return UpdateAdminProfile404JSONResponse{errNotFound("admin not found")}, nil
    }
    return UpdateAdminProfile500JSONResponse{errInternal("db error")}, nil
  }
  return UpdateAdminProfile200JSONResponse{Data: AdminProfile{Username: newUsername}}, nil
}

// --- System Info ---

func (s *Server) GetSystemInfo(ctx context.Context, _ GetSystemInfoRequestObject) (GetSystemInfoResponseObject, error) {
  return GetSystemInfo200JSONResponse{Data: SystemInfo{
    PanelVersion:   nonEmptyOrNA(buildinfo.PanelVersion),
    PanelCommitId:  nonEmptyOrNA(buildinfo.PanelCommitID),
    SingBoxVersion: nonEmptyOrNA(buildinfo.SingBoxVersion),
  }}, nil
}

// --- System Settings ---

func (s *Server) GetSystemSettings(ctx context.Context, _ GetSystemSettingsRequestObject) (GetSystemSettingsResponseObject, error) {
  subscriptionBaseURL, err := s.store.GetSystemSetting(ctx, subscriptionBaseURLKey)
  if err != nil && !errors.Is(err, db.ErrNotFound) {
    return GetSystemSettings500JSONResponse{errInternal("load settings failed")}, nil
  }

  timezone := currentSystemTimezoneName()
  savedTimezone, err := s.store.GetSystemSetting(ctx, systemTimezoneKey)
  if err != nil && !errors.Is(err, db.ErrNotFound) {
    return GetSystemSettings500JSONResponse{errInternal("load settings failed")}, nil
  }
  if err == nil {
    if normalized, _, tzErr := normalizeSystemTimezone(savedTimezone); tzErr == nil {
      timezone = normalized
    }
  }

  return GetSystemSettings200JSONResponse{Data: SystemSettings{
    SubscriptionBaseUrl: subscriptionBaseURL,
    Timezone:            timezone,
  }}, nil
}

func (s *Server) UpdateSystemSettings(ctx context.Context, request UpdateSystemSettingsRequestObject) (UpdateSystemSettingsResponseObject, error) {
  subURL := ""
  if request.Body.SubscriptionBaseUrl != nil {
    subURL = *request.Body.SubscriptionBaseUrl
  }
  tz := ""
  if request.Body.Timezone != nil {
    tz = *request.Body.Timezone
  }

  normalizedURL, err := normalizeSubscriptionBaseURL(subURL)
  if err != nil {
    return UpdateSystemSettings400JSONResponse{errBadRequest(err.Error())}, nil
  }

  normalizedTZ, _, err := normalizeSystemTimezone(tz)
  if err != nil {
    return UpdateSystemSettings400JSONResponse{errBadRequest(err.Error())}, nil
  }

  if normalizedURL == "" {
    if err := s.store.DeleteSystemSetting(ctx, subscriptionBaseURLKey); err != nil {
      return UpdateSystemSettings500JSONResponse{errInternal("save settings failed")}, nil
    }
  } else {
    if err := s.store.UpsertSystemSetting(ctx, subscriptionBaseURLKey, normalizedURL); err != nil {
      return UpdateSystemSettings500JSONResponse{errInternal("save settings failed")}, nil
    }
  }

  if err := s.store.UpsertSystemSetting(ctx, systemTimezoneKey, normalizedTZ); err != nil {
    return UpdateSystemSettings500JSONResponse{errInternal("save settings failed")}, nil
  }

  if _, err := setSystemTimezone(normalizedTZ); err != nil {
    return UpdateSystemSettings500JSONResponse{errInternal("apply timezone failed")}, nil
  }

  return UpdateSystemSettings200JSONResponse{Data: SystemSettings{
    SubscriptionBaseUrl: normalizedURL,
    Timezone:            normalizedTZ,
  }}, nil
}

// --- Sync Jobs ---

func (s *Server) ListSyncJobs(ctx context.Context, request ListSyncJobsRequestObject) (ListSyncJobsResponseObject, error) {
  limit, offset, err := paginationDefaults(request.Params.Limit, request.Params.Offset)
  if err != nil {
    return ListSyncJobs400JSONResponse{errBadRequest("invalid pagination")}, nil
  }

  filter := db.SyncJobsListFilter{Limit: limit, Offset: offset}

  if request.Params.NodeId != nil {
    filter.NodeID = *request.Params.NodeId
  }
  if request.Params.Status != nil {
    switch *request.Params.Status {
    case Queued, Running, Success, Failed:
      filter.Status = string(*request.Params.Status)
    default:
      return ListSyncJobs400JSONResponse{errBadRequest("invalid status")}, nil
    }
  }
  if request.Params.TriggerSource != nil {
    filter.TriggerSource = *request.Params.TriggerSource
  }
  if request.Params.From != nil {
    filter.From = request.Params.From
  }
  if request.Params.To != nil {
    filter.To = request.Params.To
  }

  items, err := s.store.ListSyncJobs(ctx, filter)
  if err != nil {
    return ListSyncJobs500JSONResponse{errInternal("list sync jobs failed")}, nil
  }

  out := make([]SyncJobListItem, 0, len(items))
  for _, item := range items {
    out = append(out, dbSyncJobToAPI(item))
  }
  return ListSyncJobs200JSONResponse{Data: out}, nil
}

func (s *Server) GetSyncJob(ctx context.Context, request GetSyncJobRequestObject) (GetSyncJobResponseObject, error) {
  job, err := s.store.GetSyncJobByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetSyncJob404JSONResponse{errNotFound("sync job not found")}, nil
    }
    return GetSyncJob500JSONResponse{errInternal("get sync job failed")}, nil
  }

  attempts, err := s.store.ListSyncAttemptsByJobID(ctx, request.Id)
  if err != nil {
    return GetSyncJob500JSONResponse{errInternal("list sync attempts failed")}, nil
  }

  outAttempts := make([]SyncAttempt, 0, len(attempts))
  for _, item := range attempts {
    outAttempts = append(outAttempts, dbSyncAttemptToAPI(item))
  }

  return GetSyncJob200JSONResponse{Data: SyncJobDetail{
    Job:      dbSyncJobToAPI(job),
    Attempts: outAttempts,
  }}, nil
}

func (s *Server) RetrySyncJob(ctx context.Context, request RetrySyncJobRequestObject) (RetrySyncJobResponseObject, error) {
  if request.Id <= 0 {
    return RetrySyncJob400JSONResponse{errBadRequest("invalid id")}, nil
  }

  parentJob, err := s.store.GetSyncJobByID(ctx, request.Id)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return RetrySyncJob404JSONResponse{errNotFound("sync job not found")}, nil
    }
    return RetrySyncJob500JSONResponse{errInternal("get sync job failed")}, nil
  }

  nodeItem, err := s.store.GetNodeByID(ctx, parentJob.NodeID)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return RetrySyncJob404JSONResponse{errNotFound("node not found")}, nil
    }
    return RetrySyncJob500JSONResponse{errInternal("get node failed")}, nil
  }

  parentID := parentJob.ID
  result := trySyncNodeWithSourceAndParent(ctx, s.store, nodeItem, triggerSourceRetry, &parentID)
  if result.Status != "ok" {
    return RetrySyncJob502JSONResponse{errBadGateway(result.Error)}, nil
  }

  latestJobs, err := s.store.ListSyncJobs(ctx, db.SyncJobsListFilter{
    Limit:  1,
    Offset: 0,
    NodeID: parentJob.NodeID,
  })
  if err != nil || len(latestJobs) == 0 {
    status := "ok"
    return RetrySyncJob200JSONResponse{Status: &status}, nil
  }

  dto := dbSyncJobToAPI(latestJobs[0])
  return RetrySyncJob200JSONResponse{Data: &dto}, nil
}

// --- SingBox Tools ---

func (s *Server) FormatSingBox(ctx context.Context, request FormatSingBoxRequestObject) (FormatSingBoxResponseObject, error) {
  mode := ""
  if request.Body.Mode != nil {
    mode = *request.Body.Mode
  }
  wrapped, err := wrapConfigIfNeeded(request.Body.Config, mode)
  if err != nil {
    return FormatSingBox400JSONResponse{errBadRequest(err.Error())}, nil
  }

  formatted, err := s.singBoxTools.Format(ctx, wrapped)
  if err != nil {
    return FormatSingBox400JSONResponse{errBadRequest(strings.TrimSpace(err.Error()))}, nil
  }
  return FormatSingBox200JSONResponse{Data: struct {
    Formatted string `json:"formatted"`
  }{Formatted: formatted}}, nil
}

func (s *Server) CheckSingBox(ctx context.Context, request CheckSingBoxRequestObject) (CheckSingBoxResponseObject, error) {
  mode := ""
  if request.Body.Mode != nil {
    mode = *request.Body.Mode
  }
  wrapped, err := wrapConfigIfNeeded(request.Body.Config, mode)
  if err != nil {
    return CheckSingBox400JSONResponse{errBadRequest(err.Error())}, nil
  }

  output, err := s.singBoxTools.Check(ctx, wrapped)
  if err != nil {
    return CheckSingBox200JSONResponse{Data: struct {
      Ok     bool   `json:"ok"`
      Output string `json:"output"`
    }{Ok: false, Output: strings.TrimSpace(err.Error())}}, nil
  }
  return CheckSingBox200JSONResponse{Data: struct {
    Ok     bool   `json:"ok"`
    Output string `json:"output"`
  }{Ok: true, Output: output}}, nil
}

func (s *Server) GenerateSingBox(ctx context.Context, request GenerateSingBoxRequestObject) (GenerateSingBoxResponseObject, error) {
  output, err := s.singBoxTools.Generate(ctx, request.Body.Command)
  if err != nil {
    if errors.Is(err, singboxcli.ErrInvalidGenerateKind) {
      return GenerateSingBox400JSONResponse{errBadRequest("invalid generate command")}, nil
    }
    return GenerateSingBox400JSONResponse{errBadRequest(strings.TrimSpace(err.Error()))}, nil
  }
  return GenerateSingBox200JSONResponse{Data: struct {
    Output string `json:"output"`
  }{Output: output}}, nil
}

// --- Subscription ---

func (s *Server) GetSubscription(ctx context.Context, request GetSubscriptionRequestObject) (GetSubscriptionResponseObject, error) {
  userUUID := strings.TrimSpace(request.UserUuid)
  if userUUID == "" {
    return GetSubscription404JSONResponse{errNotFound("user not found")}, nil
  }

  user, err := s.store.GetUserByUUID(ctx, userUUID)
  if err != nil {
    if errors.Is(err, db.ErrNotFound) {
      return GetSubscription404JSONResponse{errNotFound("user not found")}, nil
    }
    return GetSubscription500JSONResponse{errInternal("get user failed")}, nil
  }
  if !userstate.IsSubscriptionEligible(user, time.Now().UTC()) {
    return GetSubscription404JSONResponse{errNotFound("user not found")}, nil
  }

  inbounds, err := s.store.ListUserInbounds(ctx, user.ID)
  if err != nil {
    return GetSubscription500JSONResponse{errInternal("list inbounds failed")}, nil
  }
  items := make([]subscription.Item, 0, len(inbounds))
  for _, inb := range inbounds {
    items = append(items, subscription.Item{
      InboundUUID:       inb.InboundUUID,
      InboundType:       inb.InboundType,
      InboundTag:        inb.InboundTag,
      NodePublicAddress: inb.NodePublicAddress,
      InboundListenPort: inb.InboundListenPort,
      InboundPublicPort: inb.InboundPublicPort,
      Settings:          inb.Settings,
      TLSSettings:       inb.TLSSettings,
      TransportSettings: inb.TransportSettings,
    })
  }

  subUser := subscription.User{UUID: user.UUID, Username: user.Username}

  format := ""
  if request.Params.Format != nil {
    format = string(*request.Params.Format)
  }

  // Validate format param.
  if format != "" && format != "singbox" && format != "v2ray" {
    return GetSubscription400JSONResponse{errBadRequest("invalid format")}, nil
  }

  // Auto-detect format from User-Agent when not specified.
  if format == "" {
    ua := ""
    if ginCtx, ok := ctx.(*gin.Context); ok {
      ua = ginCtx.GetHeader("User-Agent")
    }
    if isSingboxUA(ua) {
      format = "singbox"
    } else {
      format = "v2ray"
    }
  }

  if format == "singbox" {
    payload, err := subscription.BuildSingbox(subUser, items)
    if err != nil {
      return GetSubscription500JSONResponse{errInternal("build subscription failed")}, nil
    }
    // Return raw JSON map
    var result map[string]interface{}
    if jsonErr := json.Unmarshal(payload, &result); jsonErr != nil {
      return GetSubscription500JSONResponse{errInternal("build subscription failed")}, nil
    }
    return GetSubscription200JSONResponse(result), nil
  }

  if format == "v2ray" {
    payload, err := subscription.BuildV2Ray(subUser, items)
    if err != nil {
      return GetSubscription500JSONResponse{errInternal("build subscription failed")}, nil
    }
    return GetSubscription200TextResponse(string(payload)), nil
  }

  // format was already resolved above (singbox or v2ray),
  // so this is unreachable, but return a safe default.
  return GetSubscription400JSONResponse{errBadRequest("invalid format")}, nil
}

// --- Traffic ---

func (s *Server) GetTrafficNodesSummary(ctx context.Context, request GetTrafficNodesSummaryRequestObject) (GetTrafficNodesSummaryResponseObject, error) {
  if !s.storeReady() {
    return GetTrafficNodesSummary500JSONResponse{errInternal("store not ready")}, nil
  }
  window := parseWindowParam(request.Params.Window)
  if window < 0 {
    return GetTrafficNodesSummary400JSONResponse{errBadRequest("invalid window")}, nil
  }

  p := traffic.NewSQLiteProvider(s.store)
  items, err := p.NodesSummary(ctx, window)
  if err != nil {
    return GetTrafficNodesSummary500JSONResponse{errInternal(err.Error())}, nil
  }

  out := make([]TrafficNodeSummary, 0, len(items))
  for _, it := range items {
    out = append(out, TrafficNodeSummary{
      NodeId:         it.NodeID,
      Upload:         it.Upload,
      Download:       it.Download,
      LastRecordedAt: timeRFC3339OrEmpty(it.LastRecordedAt),
      Samples:        it.Samples,
      Inbounds:       it.Inbounds,
    })
  }
  return GetTrafficNodesSummary200JSONResponse{Data: out}, nil
}

func (s *Server) GetTrafficTotalSummary(ctx context.Context, request GetTrafficTotalSummaryRequestObject) (GetTrafficTotalSummaryResponseObject, error) {
  if !s.storeReady() {
    return GetTrafficTotalSummary500JSONResponse{errInternal("store not ready")}, nil
  }
  window := parseWindowParam(request.Params.Window)
  if window < 0 {
    return GetTrafficTotalSummary400JSONResponse{errBadRequest("invalid window")}, nil
  }

  p := traffic.NewSQLiteProvider(s.store)
  it, err := p.TotalSummary(ctx, window)
  if err != nil {
    return GetTrafficTotalSummary500JSONResponse{errInternal(err.Error())}, nil
  }

  return GetTrafficTotalSummary200JSONResponse{Data: TrafficTotalSummary{
    Upload:         it.Upload,
    Download:       it.Download,
    LastRecordedAt: timeRFC3339OrEmpty(it.LastRecordedAt),
    Samples:        it.Samples,
    Nodes:          it.Nodes,
    Inbounds:       it.Inbounds,
  }}, nil
}

func (s *Server) GetTrafficTimeseries(ctx context.Context, request GetTrafficTimeseriesRequestObject) (GetTrafficTimeseriesResponseObject, error) {
  if !s.storeReady() {
    return GetTrafficTimeseries500JSONResponse{errInternal("store not ready")}, nil
  }
  window := parseWindowParam(request.Params.Window)
  if window < 0 {
    return GetTrafficTimeseries400JSONResponse{errBadRequest("invalid window")}, nil
  }

  bucket := traffic.BucketHour
  if request.Params.Bucket != nil {
    switch *request.Params.Bucket {
    case Minute:
      bucket = traffic.BucketMinute
    case Hour:
      bucket = traffic.BucketHour
    case Day:
      bucket = traffic.BucketDay
    default:
      return GetTrafficTimeseries400JSONResponse{errBadRequest("invalid bucket")}, nil
    }
  }

  var nodeID int64
  if request.Params.NodeId != nil {
    nodeID = *request.Params.NodeId
  }

  p := traffic.NewSQLiteProvider(s.store)
  items, err := p.Timeseries(ctx, traffic.TimeseriesQuery{
    Window: window,
    Bucket: bucket,
    NodeID: nodeID,
  })
  if err != nil {
    return GetTrafficTimeseries500JSONResponse{errInternal(err.Error())}, nil
  }

  out := make([]TrafficTimeseriesPoint, 0, len(items))
  for _, it := range items {
    out = append(out, TrafficTimeseriesPoint{
      BucketStart: timeRFC3339OrEmpty(it.BucketStart),
      Upload:      it.Upload,
      Download:    it.Download,
    })
  }
  return GetTrafficTimeseries200JSONResponse{Data: out}, nil
}

// parseWindowParam converts the optional window string to a duration.
func parseWindowParam(w *string) time.Duration {
  if w == nil || *w == "" {
    return 24 * time.Hour
  }
  return parseWindowOrDefault(*w, 24*time.Hour)
}

// --- Helpers for sync job DTOs ---

func dbSyncJobToAPI(item db.SyncJob) SyncJobListItem {
  out := SyncJobListItem{
    Id:              item.ID,
    NodeId:          item.NodeID,
    ParentJobId:     item.ParentJobID,
    TriggerSource:   item.TriggerSource,
    Status:          item.Status,
    InboundCount:    item.InboundCount,
    ActiveUserCount: item.ActiveUserCount,
    PayloadHash:     item.PayloadHash,
    AttemptCount:    item.AttemptCount,
    DurationMs:      item.DurationMS,
    ErrorSummary:    item.ErrorSummary,
    CreatedAt:       item.CreatedAt,
  }
  if item.StartedAt != nil {
    out.StartedAt = item.StartedAt
  }
  if item.FinishedAt != nil {
    out.FinishedAt = item.FinishedAt
  }
  return out
}

func dbSyncAttemptToAPI(item db.SyncAttempt) SyncAttempt {
  out := SyncAttempt{
    Id:           item.ID,
    AttemptNo:    item.AttemptNo,
    Status:       item.Status,
    HttpStatus:   item.HTTPStatus,
    DurationMs:   item.DurationMS,
    ErrorSummary: item.ErrorSummary,
    BackoffMs:    item.BackoffMS,
    StartedAt:    item.StartedAt,
  }
  if item.FinishedAt != nil {
    out.FinishedAt = item.FinishedAt
  }
  return out
}
