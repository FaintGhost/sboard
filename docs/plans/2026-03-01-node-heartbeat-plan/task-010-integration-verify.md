# Task 010: 集成验证

**depends-on:** task-005-heartbeat-handler-impl, task-006-node-heartbeat-impl, task-007-approve-reject-impl, task-008-monitor-impl, task-009-frontend-pending-ui

## Summary

全面验证所有组件协同工作。运行完整测试套件，确保无回归。

## BDD Scenario

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

## Verification Steps

1. **前端编译 + 类型检查:**
   ```bash
   cd web && bunx tsc -b
   ```

2. **前端单元测试:**
   ```bash
   cd web && bun run test
   ```

3. **Panel 全量测试:**
   ```bash
   cd panel && go test ./... -count=1
   ```

4. **Node 全量测试:**
   ```bash
   cd node && go test ./... -count=1
   ```

5. **Proto 生成检查:**
   ```bash
   moon run panel:check-generate
   ```

6. **现有 RPC auth 测试不受影响:**
   ```bash
   cd panel && go test ./internal/rpc/ -run TestRPCAuth -v
   ```

7. **现有 REST auth 测试不受影响:**
   ```bash
   cd panel && go test ./internal/api/ -run TestAuth -v
   ```

## Success Criteria

- 所有测试通过
- 无类型错误
- Proto 生成产物与源码一致
- 现有功能无回归
