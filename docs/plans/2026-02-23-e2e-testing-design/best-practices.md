# E2E 测试最佳实践

## Docker 相关

### 镜像版本锁定

- Playwright Docker 镜像版本 **必须** 与 `@playwright/test` npm 包版本完全匹配
- 在 `package.json` 和 `Dockerfile` 中同步更新版本
- 不使用 `:latest` 标签

### IPC 配置

```yaml
playwright:
  ipc: host  # Chromium 需要共享内存，否则可能 OOM 崩溃
```

### Healthcheck 优于 wait-for-it

- 使用 Docker 原生 `healthcheck` + `depends_on: condition: service_healthy`
- 不使用 `wait-for-it.sh`（只检查端口，不检查应用就绪状态）
- `start_period` 给服务充足的启动缓冲时间

### 临时卷管理

```bash
# 每次测试前确保清理
docker compose -f docker-compose.e2e.yml down -v
# 运行测试
docker compose -f docker-compose.e2e.yml up --build ...
```

## Playwright 相关

### 选择器策略

优先级（从高到低）：
1. `getByRole()` — 语义化选择器，最稳定
2. `getByText()` — 文本内容选择
3. `getByTestId()` — data-testid 属性（需要前端配合添加）
4. `getByLabel()` — 表单标签
5. CSS 选择器 — 最后手段

```typescript
// 推荐
await page.getByRole('button', { name: '创建用户' }).click();
await page.getByRole('textbox', { name: '用户名' }).fill('testuser');

// 避免
await page.click('.btn-primary');
await page.locator('#username-input').fill('testuser');
```

### 自动等待

Playwright 内置自动等待机制，不要使用硬编码 sleep：

```typescript
// 推荐 — Playwright 自动等待元素可见且可交互
await page.getByRole('button', { name: '保存' }).click();
await expect(page.getByText('保存成功')).toBeVisible();

// 避免
await page.waitForTimeout(2000);
await page.click('button');
```

### 测试隔离

```typescript
// 每个测试文件使用独立的测试数据
test('创建用户', async ({ authenticatedPage }) => {
  const uniqueName = `user-${Date.now()}`;
  // ... 使用 uniqueName 创建用户
});
```

### 失败处理

```typescript
// playwright.config.ts
{
  screenshot: 'only-on-failure',       // 失败时自动截图
  video: 'retain-on-failure',          // 失败时保留视频
  trace: isCI ? 'on-first-retry' : 'off', // CI 首次重试时收集 trace
}
```

### API 与 UI 测试结合

对于 Node 配置验证等场景，直接使用 Playwright 的 `request` API 而非浏览器操作：

```typescript
// 通过 API 验证 Node 配置
const response = await request.get(`${NODE_API_URL}/api/health`);
expect(response.status()).toBe(200);
```

## 代码组织

### Fixture 复用

将通用的 setup 逻辑（认证、数据创建）封装为 Playwright fixtures，避免每个测试文件重复。

### Page Object 模式（可选）

如果页面交互复杂，可以引入 Page Object：

```typescript
// 但在初始阶段，简单的 fixture + 直接操作足够
// 不要过早引入 Page Object 增加复杂度
```

### 测试命名

```typescript
// 使用描述性的中文测试名
test('创建新用户后在列表中可见', async ({ page }) => { });
test('删除用户后列表更新', async ({ page }) => { });
test('节点同步后配置生效', async ({ page }) => { });
```

## CI 集成（后续扩展）

### GitHub Actions 示例结构

```yaml
# .github/workflows/e2e.yml
name: E2E Tests
on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run E2E tests
        run: make e2e
      - name: Upload report
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: e2e/playwright-report/
          retention-days: 30
```

## 常见陷阱

| 陷阱 | 解决方案 |
|------|---------|
| Playwright 镜像版本与 npm 包不匹配 | 两处版本保持同步 |
| 浏览器 OOM | 设置 `ipc: host` |
| 测试间数据污染 | 使用临时卷 + 唯一名称前缀 |
| Healthcheck 中 curl 不存在 | Panel/Node 的 Debian 镜像需安装 curl |
| 网络不通 | 确保所有服务在同一 Docker network |
| Token 过期 | 测试 setup 中每次重新获取 token |
| 前端路由 404 | Panel 配置 `PANEL_SERVE_WEB=true` 且 SPA fallback 正确 |

## 性能优化

1. **并行构建**：Panel 和 Node 的 Docker 镜像可并行构建
2. **Layer 缓存**：利用 Docker BuildKit cache mount 加速 Go 编译和前端打包
3. **浏览器复用**：Playwright 在同一 project 内复用浏览器实例
4. **按需运行**：通过 `--project=smoke` 只运行快速测试
