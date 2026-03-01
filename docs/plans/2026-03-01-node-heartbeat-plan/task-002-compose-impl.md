# Task 002: Docker Compose 模板 — 实现

**depends-on:** task-002-compose-test

## Summary

修改 `buildNodeDockerCompose` 函数，支持 `panelUrl` 和 `nodeUuid` 参数，在生成的 docker-compose.yml 中输出对应环境变量。

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

- **Modify:** `web/src/lib/node-compose.ts`
  - `BuildNodeComposeInput` 类型新增 `panelUrl?: string` 和 `nodeUuid?: string` 字段
  - 在 `buildNodeDockerCompose` 函数中，当 `panelUrl` 非空时输出 `PANEL_URL` 环境变量
  - 当 `nodeUuid` 非空时输出 `NODE_UUID` 环境变量
  - 对这两个值也使用 `escapeDoubleQuoted` 转义

## Steps

1. 扩展 `BuildNodeComposeInput` 类型
2. 在 YAML 模板字符串中追加条件行
3. 运行测试验证 GREEN

## Verify

```bash
cd web && bun run test -- node-compose
```

所有测试应该 PASS。
