# Task 003: Node UUID 持久化 — 测试

**depends-on:** (none)

## Summary

为 Node UUID 的生成、持久化和加载功能编写测试。UUID 在首次启动时生成，写入磁盘，后续启动读取已有 UUID。

## BDD Scenario

```gherkin
Scenario: Node generates UUID on first start
  Given no UUID file exists at the configured path
  When Node loads or generates UUID
  Then a valid UUIDv4 is generated
  And the UUID is written to the file with 0600 permissions

Scenario: Node loads existing UUID on restart
  Given a UUID file exists containing "existing-uuid-value"
  When Node loads or generates UUID
  Then it returns "existing-uuid-value" without generating a new one

Scenario: Node regenerates UUID if file is corrupted
  Given a UUID file exists but contains empty/whitespace content
  When Node loads or generates UUID
  Then a new valid UUIDv4 is generated and persisted
```

## Files

- **Create:** `node/internal/config/uuid_test.go`
  - `TestLoadOrGenerateUUID_NewFile` — 无文件时生成新 UUID
  - `TestLoadOrGenerateUUID_ExistingFile` — 已有文件时加载
  - `TestLoadOrGenerateUUID_CorruptedFile` — 损坏文件时重新生成
  - 使用 `t.TempDir()` 隔离测试文件系统

## Steps

1. 创建测试文件，import `testing` 和 `os`
2. 每个测试用 `t.TempDir()` 创建临时目录
3. 验证 UUID 格式（`google/uuid` 的 `Parse` 函数）
4. 验证文件权限（`os.Stat` 检查 `0600`）

## Verify

```bash
cd node && go test ./internal/config/ -run TestLoadOrGenerateUUID -v
```

应该 FAIL（函数不存在）。
