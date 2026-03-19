# Inbound 模板 `settings.users` 为空设计说明

## 目标

明确 Panel 管理面里的 inbound 模板为什么允许 `settings.users` 为空，并固定当前实现的边界：

- 管理面模板负责保存静态协议配置。
- 运行态下发给 Node 的 `users` 列表由 Panel sync 阶段按分组动态注入。
- 协议校验只校验“如果提供了用户项，这些用户项是否合法”，而不是强制模板阶段必须先写死用户列表。

## 背景

当前创建和更新 inbound 时，RPC 服务会先把 `settings_json` 解析成 map，再调用 `inbounds.ValidateSettings(protocol, settingsMap)` 做模板校验。这里要求 `settings` 本身必须是合法 JSON 对象，但并不要求 `settings.users` 必须存在或非空。

随后在 Panel 向 Node 下发配置时，`BuildSyncPayload` 会为每个 inbound 重新构造运行态 payload，并统一调用 `buildUsersForProtocol` 按分组用户生成 `users` 字段。

这意味着：

- 模板阶段的 `settings.users` 不是运行态事实源。
- 运行态的 `users` 列表来自分组用户集，而不是来自管理员在模板里手填的静态数组。

## 设计结论

结论只有一条：

Panel 下发配置给 Node 时，会把分组用户动态注入到每个 inbound 的 `users` 字段，因此 inbound 模板里的 `settings.users` 可以为空。

换句话说，模板和运行态承担的是不同职责：

- 模板负责保存协议、端口、TLS、传输等静态骨架。
- sync 负责把该分组下当前有效的用户集映射成协议所需的 `users` 结构。

## 当前实现边界

### 1. 模板校验边界

`ValidateSettings` 的行为是“按协议校验 settings 中已经存在的字段”，不是“补全运行态字段”。

因此：

- `settings.users` 缺失时，多数协议校验会直接通过。
- `settings.users` 为空数组时，也会通过。
- 如果 `settings.users` 里已经提供了用户项，则继续逐项校验字段格式，例如：
  - `vless`/`vmess` 校验 `uuid`
  - `trojan`/`hysteria2`/`shadowtls`/`anytls` 校验 `password`
  - `socks`/`http`/`mixed`/`naive` 校验 `username` 与 `password`
  - `tuic` 校验 `uuid` 与 `password`

这保证了模板阶段既允许空用户列表，也不会放过显式写入的非法用户项。

### 2. 运行态注入边界

`BuildSyncPayload` 在组装 Node 配置时，总是为每个 inbound 设置 `item["users"]`，其值来自 `buildUsersForProtocol(protocol, users, settings)`。

这一步使用的是 sync 时计算出来的分组用户集，而不是数据库里原始保存的模板 `settings.users`。

因此运行态的 `users` 语义是：

- 与分组绑定
- 与用户状态绑定
- 与协议字段映射绑定
- 每次 sync 重新计算

### 3. 模板字段与运行态字段的区别

某些字段虽然写在模板 settings 中，但真正生效的位置在运行态用户项里，而不是 inbound 顶层。例如：

- `vless` 的 `flow` 在 sync 时只落到用户级字段，不落到 inbound 顶层。
- `shadowsocks 2022` 会在 sync 时为 inbound 顶层补服务端 `password`，并为每个用户生成独立 `users[].password`。

这类行为进一步说明：管理面模板不等价于最终下发给 Node 的完整运行态结构。

## 协议映射清单

当前 `buildUsersForProtocol` 已覆盖下列协议，sync 时会自动生成对应的 `users` 结构：

- `vless`：`name`、`uuid`，如配置了 `flow`，额外写入用户级 `flow`
- `vmess`：`name`、`uuid`
- `trojan`：`name`、`password`
- `shadowsocks`：`name`、`password`
- `socks`：`username`、`password`
- `http`：`username`、`password`
- `mixed`：`username`、`password`
- `hysteria2`：`name`、`password`
- `tuic`：`name`、`uuid`、`password`
- `naive`：`username`、`password`
- `shadowtls`：`name`、`password`
- `anytls`：`name`、`password`

其中：

- `socks`、`http`、`mixed` 共享同一套 auth users 生成逻辑。
- `shadowtls`、`anytls` 共享同一套用户密码映射逻辑。
- `shadowsocks 2022` 在生成用户密码时不是直接复用 UUID，而是做确定性派生。

## 为什么这不是“放松校验过头”

允许模板中的 `settings.users` 为空，并不意味着系统不校验用户配置，而是把校验拆成了两个层次：

- 模板层：校验协议骨架和显式提供的用户项是否合法。
- sync 层：根据真实分组用户集生成运行态 `users`。

这样做的原因是当前系统的数据模型本来就是“用户属于 group，Node 也属于 group，Panel 在 sync 时把 group 用户注入 inbound”。如果强制模板阶段必须填写 `settings.users`，反而会制造两份来源：

- 一份是模板里人工维护的静态 users
- 一份是分组关系计算出来的动态 users

两份来源一旦不一致，就会让最终运行态到底信哪一份变得不清楚。

当前实现选择单一事实源：运行态用户集只信 sync 阶段按 group 计算出的结果。

## 与现有测试的一致性

当前行为已经被两类测试覆盖：

- `validators_test.go` 验证多个协议在 `missing users` 或 `empty users` 时允许通过
- `build_config.go` 相关测试验证 sync 阶段会按协议把用户注入到最终 payload 中

这两类测试组合起来，表达的就是当前约定：

- 模板可以空
- 下发时补齐

## 维护规则

后续如果新增协议或调整协议字段映射，必须同时满足以下规则：

1. 新协议若依赖运行态用户注入，必须在 `buildUsersForProtocol` 中显式定义字段映射。
2. 新协议若允许模板 `settings.users` 为空，必须在对应 validator 中保持“空数组/缺失字段可通过，但已有用户项要逐项校验”的行为。
3. 新协议必须同时补两类测试：
  - validator 对 `missing users`/`empty users` 的测试
  - sync payload 对运行态用户注入结果的测试

如果将来某个协议确实要求“模板阶段就必须显式提供 users”，那应该作为特例在 validator 中单独声明，并在设计文档里明确说明为什么它不能沿用当前的 group 注入模型。

## 非目标

这份说明不改变以下事实：

- `settings` 本身仍然必须是合法 JSON 对象
- 各协议自身的其他必填字段仍然照常校验
- sync 时注入的仍然是该 group 下当前有效、可下发的用户集
- 这不是在支持“模板 users 与 group users 双来源并存”

本设计只是在文档层面把当前已经实现的规则明确化，避免后续把模板校验和运行态注入混为一谈。