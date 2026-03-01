# BDD Specifications — Node→Panel 心跳注册

## Feature: Node 心跳注册

### Scenario 1: Node 配置 PANEL_URL 后定期心跳

```gherkin
Feature: Node heartbeat to Panel

  Scenario: Node sends heartbeat when PANEL_URL is configured
    Given a Node running with PANEL_URL="http://panel:8080" and NODE_UUID="uuid-1"
    And the Node has SECRET_KEY="key-1"
    When 30 seconds elapse
    Then Node sends a Heartbeat RPC to Panel with uuid="uuid-1", secret_key="key-1"
    And the heartbeat includes Node version and api_addr

  Scenario: Node does NOT send heartbeat when PANEL_URL is empty
    Given a Node running without PANEL_URL configured
    When 60 seconds elapse
    Then no Heartbeat RPC is sent
    And Node operates in passive-only mode (current behavior)
```

### Scenario 2: Panel 识别已知 Node

```gherkin
  Scenario: Panel receives heartbeat from known Node
    Given Panel has a Node record with uuid="uuid-1" and secret_key="key-1"
    When Node sends Heartbeat with uuid="uuid-1" and secret_key="key-1"
    Then Panel updates last_seen_at for that Node
    And responds with status=RECOGNIZED

  Scenario: Panel rejects heartbeat with mismatched secret_key
    Given Panel has a Node record with uuid="uuid-1" and secret_key="key-1"
    When Node sends Heartbeat with uuid="uuid-1" and secret_key="wrong-key"
    Then Panel responds with status=REJECTED
    And does NOT update any Node record
```

### Scenario 3: Panel 创建待认领节点

```gherkin
  Scenario: Panel receives heartbeat from unknown Node (after reset)
    Given Panel has no Node record with uuid="uuid-1"
    When Node sends Heartbeat with uuid="uuid-1" and secret_key="key-1" and api_addr=":3003"
    Then Panel creates a pending Node record with status="pending"
    And responds with status=PENDING

  Scenario: Duplicate heartbeat from pending Node
    Given Panel has a pending Node with uuid="uuid-1"
    When Node sends another Heartbeat with uuid="uuid-1"
    Then Panel updates last_seen_at on the pending record
    And responds with status=PENDING (no duplicate record created)
```

### Scenario 4: 管理员审批待认领节点

```gherkin
  Scenario: Admin approves a pending Node
    Given Panel has a pending Node with uuid="uuid-1" and api_address="1.2.3.4" and api_port=3003
    When admin calls ApproveNode with name="US-West-1" and group_id=1
    Then Node status changes from "pending" to "offline"
    And Node name is set to "US-West-1"
    And Node group_id is set to 1
    And NodesMonitor will detect the Node as online on next check
    And SyncConfig is triggered automatically

  Scenario: Admin rejects a pending Node
    Given Panel has a pending Node with uuid="uuid-1"
    When admin calls RejectNode for uuid="uuid-1"
    Then the pending Node record is deleted
```

### Scenario 5: 完整 Panel 重置恢复流程

```gherkin
  Scenario: Full Panel reset recovery
    Given a Panel with Node "US-1" (uuid="uuid-1", address="1.2.3.4:3003")
    And Node is online and serving users
    When admin deletes Panel data folder and re-bootstraps
    Then Panel has no Node records
    And Node continues serving users with last synced config
    And within 30 seconds Node sends heartbeat to Panel
    And Panel creates a pending record for uuid="uuid-1"
    And admin sees "1 pending node" in the UI
    When admin approves the pending node with name="US-1"
    Then Panel pushes fresh config to Node
    And Node updates its config (existing users remain if still in new config)
```

### Scenario 6: Docker Compose 模板生成

```gherkin
  Scenario: Docker Compose includes PANEL_URL
    Given Panel is running at "http://panel.example.com:8080"
    When admin creates a new Node via UI
    Then the generated docker-compose.yml includes PANEL_URL environment variable
    And includes NODE_UUID environment variable
    And includes all existing variables (NODE_HTTP_ADDR, NODE_SECRET_KEY, NODE_LOG_LEVEL)
```

---

## Testing Strategy

### Unit Tests

| Component | Test Focus | Location |
|-----------|-----------|----------|
| Node heartbeat goroutine | Timer, retry, graceful shutdown | `node/internal/heartbeat/heartbeat_test.go` |
| Node UUID persistence | Generate, persist, reload | `node/internal/config/uuid_test.go` |
| Panel Heartbeat handler | Known/unknown/pending/rejected logic | `panel/internal/rpc/node_registration_test.go` |
| Panel ApproveNode | Status transition, SyncConfig trigger | `panel/internal/rpc/node_registration_test.go` |
| Docker Compose template | PANEL_URL and NODE_UUID in output | `web/src/lib/node-compose.test.ts` |

### Integration Tests

| Scenario | Test Focus | Location |
|----------|-----------|----------|
| Node↔Panel heartbeat round-trip | Full RPC call with httptest | `panel/internal/rpc/node_registration_test.go` |
| Pending→Approved→Synced flow | Complete lifecycle | `panel/internal/api/node_approval_test.go` |

### E2E Tests

| Scenario | Test Focus | Location |
|----------|-----------|----------|
| Panel reset recovery | Full Docker Compose scenario | `e2e/tests/e2e/node-heartbeat.spec.ts` |
