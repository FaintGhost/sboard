# BDD Specifications

## Feature: Panel-Node RPC 同步链路

```gherkin
Feature: Panel 通过 RPC 向 Node 下发配置
  作为管理员
  我希望 Panel 与 Node 通过 RPC 同步配置
  以替代旧 REST 通信并保持同步任务可追踪

  Scenario: 同步成功
    Given Panel 与 Node 的 RPC 服务均已启动
    And Node 鉴权密钥与 Panel 中配置一致
    When 管理员触发节点同步
    Then Node 成功应用配置
    And SyncJob 状态为 success

  Scenario: 鉴权失败
    Given Node 鉴权密钥与 Panel 配置不一致
    When 管理员触发节点同步
    Then 返回 unauthenticated 错误
    And SyncJob 状态为 failed
    And 错误摘要可用于定位密钥问题

  Scenario: 配置非法
    Given 入站配置包含非法字段
    When 管理员触发节点同步
    Then 返回 invalid_argument 错误
    And SyncJob 状态为 failed
    And 不进入重试

  Scenario: 节点不可达
    Given Node 网络不可达
    When 管理员触发节点同步
    Then 返回 unavailable 或 deadline_exceeded
    And SyncJob 状态为 failed
```

## Feature: 监控与流量采样迁移到 RPC

```gherkin
Feature: 监控链路改走 RPC
  作为系统运维
  我希望健康检查和流量采样通过 RPC 获取
  以统一通信协议并减少维护面

  Scenario: 节点健康检查
    Given Node RPC Health 可访问
    When Panel monitor 执行健康探测
    Then 节点状态被正确更新为 online

  Scenario: 获取网卡流量
    Given Node 正常运行
    When Panel monitor 拉取 traffic 样本
    Then 返回包含 rx_bytes 和 tx_bytes 的结果

  Scenario: 获取入站统计并重置
    Given Node 有入站流量数据
    When Panel 请求 InboundTraffic 且 reset=true
    Then 返回 data 与 meta
    And Node 端统计被按预期重置
```

## Feature: 并发与一致性

```gherkin
Feature: 同一节点并发同步保护
  作为系统
  我希望同一节点的同步串行执行
  以避免配置互相覆盖

  Scenario: 并发同步请求串行化
    Given 同一节点在短时间内收到两个同步请求
    When 两个请求并发执行
    Then 节点最终配置与最后一次成功请求一致
    And 不出现中间态损坏
```

## Feature: REST 兼容边界

```gherkin
Feature: 仅保留订阅 REST
  作为系统管理员
  我希望切换后管理与节点控制走 RPC
  且用户订阅 REST 保持可用

  Scenario: 管理 REST 路径不可用
    Given 系统已完成直切发布
    When 客户端访问历史管理 REST 路径
    Then 返回 not found 或明确不可用

  Scenario: 订阅 REST 保持可用
    Given 用户存在有效订阅
    When 客户端访问 GET /api/sub/:user_uuid?format=singbox
    Then 返回有效 sing-box 配置

  Scenario: 订阅默认格式按 User-Agent 选择
    Given 请求未显式指定 format
    When 不同 User-Agent 访问订阅接口
    Then 返回符合约定的默认格式
```

## Feature: Shadowsocks 2022 兼容

```gherkin
Feature: SS 2022 密钥格式保持一致
  作为订阅消费者
  我希望切换通信协议后 SS 2022 行为不变
  以避免客户端连通性回归

  Scenario: 同步与订阅均满足 SS 2022 密钥规则
    Given 入站协议为 shadowsocks 2022
    When Panel 生成节点下发配置与订阅
    Then 服务端 password 为有效 base64 密钥
    And 用户密码满足 psk:userKey 组合规则
```

## 自动化验证矩阵

- 单元测试：
  - Panel Node RPC 客户端错误映射与超时行为。
  - Node RPC 服务鉴权、参数校验、core 调用分支。
  - SS 2022 密钥与用户筛选规则。
- 集成测试：
  - Panel `runNodeSync` + mock Node RPC 服务。
  - Node RPC handler + `sync.ParseAndValidateConfig` 协同。
- E2E：
  - 创建用户/分组/节点/入站 -> 触发同步 -> 校验 Node 状态与流量。
  - 订阅 REST `singbox|v2ray` 与 UA fallback 场景。

## 回归门禁

- `make check-generate`
- `cd panel && go test ./... -count=1`
- `cd node && go test ./... -count=1`
- `cd panel/web && bun run lint`
- `cd panel/web && bun run format`
- `cd panel/web && bunx tsc -b`
- `cd panel/web && bun run test`
- `make e2e-smoke`
- `make e2e`
