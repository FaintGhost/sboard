# BDD 规格说明

## Smoke 测试场景

### Feature: 系统健康检查

```gherkin
Feature: 系统健康检查
  作为运维人员
  我需要快速验证 Panel 和 Node 服务可用
  以确认部署成功

  Scenario: Panel 健康检查
    When 发送 GET 请求到 /api/health
    Then 返回状态码 200
    And 响应体包含 status "ok"

  Scenario: Node 健康检查
    When 发送 GET 请求到 Node 的 /api/health
    Then 返回状态码 200
    And 响应体包含 status "ok"
```

### Feature: Bootstrap 初始化

```gherkin
Feature: Bootstrap 初始化
  作为系统管理员
  首次访问系统时需要创建管理员账户
  以便后续管理

  Scenario: 首次访问显示 Bootstrap 页面
    Given 系统是全新状态（无管理员）
    When 访问首页
    Then 应重定向到登录页
    And 显示初始化设置表单

  Scenario: 成功完成 Bootstrap
    Given 系统需要初始化
    When 在 Bootstrap 表单中填入管理员信息
      | 字段 | 值 |
      | 用户名 | admin |
      | 密码 | admin1234 |
      | 确认密码 | admin1234 |
    And 提交表单
    Then 管理员账户创建成功
    And 自动登录并跳转到 Dashboard

  Scenario: 登录已初始化的系统
    Given 系统已完成 Bootstrap
    When 访问登录页
    And 输入正确的管理员凭据
    Then 登录成功
    And 跳转到 Dashboard
```

### Feature: 核心页面导航

```gherkin
Feature: 核心页面导航
  作为已登录的管理员
  我需要验证所有核心页面可正常加载

  Scenario Outline: 核心页面加载
    Given 管理员已登录
    When 导航到 <路由>
    Then 页面正常加载（无错误弹窗或白屏）
    And 页面标题或关键元素可见

    Examples:
      | 路由 |
      | / |
      | /users |
      | /groups |
      | /nodes |
      | /inbounds |
      | /sync-jobs |
      | /subscriptions |
      | /settings |
```

## E2E 测试场景

### Feature: 认证管理

```gherkin
Feature: 认证管理
  作为管理员
  我需要安全地登录和登出系统

  Scenario: 登录成功
    Given 访问登录页
    When 输入正确的用户名和密码
    And 点击登录按钮
    Then 跳转到 Dashboard
    And 侧边栏导航可见

  Scenario: 登录失败 - 错误密码
    Given 访问登录页
    When 输入正确的用户名和错误的密码
    And 点击登录按钮
    Then 显示错误提示信息
    And 仍停留在登录页

  Scenario: 未认证访问受保护页面
    Given 未登录状态
    When 直接访问 /users
    Then 重定向到登录页
```

### Feature: 用户管理

```gherkin
Feature: 用户管理
  作为管理员
  我需要管理系统中的用户

  Scenario: 创建新用户
    Given 管理员已登录
    And 导航到用户管理页面
    When 点击"创建用户"按钮
    And 填写用户信息表单
    And 提交表单
    Then 用户创建成功
    And 用户列表中显示新创建的用户

  Scenario: 编辑用户信息
    Given 管理员已登录
    And 用户列表中有一个测试用户
    When 点击该用户的编辑按钮
    And 修改用户信息
    And 保存更改
    Then 用户信息更新成功
    And 列表中显示更新后的信息

  Scenario: 删除用户
    Given 管理员已登录
    And 用户列表中有一个测试用户
    When 点击该用户的删除按钮
    And 确认删除
    Then 用户被成功删除
    And 用户列表中不再显示该用户
```

### Feature: 分组管理

```gherkin
Feature: 分组管理
  作为管理员
  我需要通过分组来组织用户

  Scenario: 创建分组
    Given 管理员已登录
    And 导航到分组管理页面
    When 点击"创建分组"按钮
    And 填写分组名称
    And 提交表单
    Then 分组创建成功

  Scenario: 将用户分配到分组
    Given 存在一个分组和一个用户
    When 编辑用户，选择所属分组
    And 保存更改
    Then 用户的分组信息更新成功
```

### Feature: 节点管理

```gherkin
Feature: 节点管理
  作为管理员
  我需要管理代理节点

  Scenario: 创建节点
    Given 管理员已登录
    And 导航到节点管理页面
    When 点击"创建节点"按钮
    And 填写节点信息
      | 字段 | 值 |
      | 名称 | test-node |
      | API 地址 | node |
      | API 端口 | 3000 |
      | 密钥 | e2e-test-node-secret |
      | 公网地址 | node |
    And 提交表单
    Then 节点创建成功
    And 节点列表中显示新节点

  Scenario: 查看节点健康状态
    Given 存在一个已配置的节点
    When 查看节点详情或列表
    Then 节点状态显示为在线（健康检查通过）

  Scenario: 删除节点
    Given 存在一个节点
    When 删除该节点
    Then 节点被成功移除
```

### Feature: 配置同步与验证

```gherkin
Feature: 配置同步与验证
  作为管理员
  我需要将入站配置同步到节点
  并验证配置正确生效

  Scenario: 创建入站配置并同步到节点
    Given 存在一个在线的节点
    When 创建一个入站配置并关联到该节点
    And 触发节点同步
    Then 同步状态显示成功

  Scenario: 验证 Node 接收到正确配置（API 级别）
    Given 已成功同步配置到节点
    When 通过 Node API 查询当前配置
    Then 配置内容与 Panel 下发的一致
    And sing-box 进程状态为 running

  Scenario: 修改入站配置后重新同步
    Given 节点已有生效的配置
    When 在 Panel 中修改入站配置
    And 重新触发同步
    Then Node 端配置更新为最新版本
```

### Feature: 订阅管理

```gherkin
Feature: 订阅管理
  作为管理员
  我需要为用户生成订阅链接

  Scenario: 生成订阅链接
    Given 存在用户和已配置的入站
    When 查看用户的订阅信息
    Then 能看到订阅链接

  Scenario: 验证订阅内容
    Given 用户有有效的订阅链接
    When 访问订阅链接
    Then 返回有效的代理配置内容
    And 配置格式正确（sing-box JSON 或 v2ray base64）
```

## 测试数据策略

### 测试顺序依赖

```
Bootstrap (setup)
  → Auth tests
  → Group creation
  → User creation (with group)
  → Node creation
  → Inbound creation (linked to node)
  → Sync trigger + verification
  → Subscription verification
  → Cleanup (container destruction)
```

### 数据隔离

- 每次 `docker compose up` 创建全新的临时 volume
- `docker compose down -v` 销毁所有数据
- 测试间使用唯一名称前缀（如 `e2e-{timestamp}-`）避免冲突
