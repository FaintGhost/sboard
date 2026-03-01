# Task 002: Docker Compose 模板 — 测试

**depends-on:** (none)

## Summary

为 `buildNodeDockerCompose` 函数添加测试，验证生成的 docker-compose.yml 包含 `PANEL_URL` 和 `NODE_UUID` 环境变量。

## BDD Scenario

```gherkin
Scenario: Docker Compose includes PANEL_URL
  Given Panel is running at "http://panel.example.com:8080"
  When admin creates a new Node via UI
  Then the generated docker-compose.yml includes PANEL_URL environment variable
  And includes NODE_UUID environment variable
  And includes all existing variables (NODE_HTTP_ADDR, NODE_SECRET_KEY, NODE_LOG_LEVEL)
```

## Files

- **Modify:** `web/src/lib/node-compose.test.ts`
  - 新增测试: 验证输出包含 `PANEL_URL` 和 `NODE_UUID` 环境变量
  - 新增测试: 验证 `panelUrl` 为空时不输出 `PANEL_URL` 行
  - 新增测试: 验证 `nodeUuid` 在输入中传入时正确渲染
  - 保留现有测试（更新 snapshot 以匹配新字段）

## Steps

1. 在现有测试文件中添加新的 test case
2. 更新现有 snapshot 测试以包含 `panelUrl` 和 `nodeUuid` 参数
3. 测试应该先 FAIL（RED），因为实现还没改

## Verify

```bash
cd web && bun run test -- node-compose
```

应该 FAIL（实现未修改）。
