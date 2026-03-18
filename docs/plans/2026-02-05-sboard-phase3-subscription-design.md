# SBoard 阶段 3（订阅系统 - sing-box）设计

## 目标
- 新增订阅端点，用户通过唯一订阅链接获取与其关联节点的 outbounds。
- 仅实现 sing-box 输出，按 User-Agent 自动返回 JSON 或 Base64(JSON)。
- 支持节点对外地址与 Node API 地址分离，支持入站对外端口覆盖。

## 关键决策
- 订阅端点：`GET /api/sub/:user_uuid`
- `?format=` 优先于 UA，当前仅支持 `format=singbox`
- UA 命中 `sing-box`/`SFA`/`SFI` 返回 JSON，否则返回 Base64(JSON)
- 用户非 `active`（`disabled/expired/traffic_exceeded`）或不存在 → 404
- 订阅内容仅包含 outbounds，不包含 `dns/route`
- 用户按节点分配：新增 `user_nodes` 表

## 数据模型调整
### nodes
新增字段：
- `api_address`（Panel ↔ Node 通信用）
- `api_port`
- `public_address`（订阅用）

### inbounds
新增字段：
- `public_port`（订阅用，优先于 `listen_port`）

### user_nodes
新表：
```
user_nodes (
  user_id INTEGER REFERENCES users(id),
  node_id INTEGER REFERENCES nodes(id),
  PRIMARY KEY (user_id, node_id)
)
```

## 订阅生成规则
1. 通过 `user_uuid` 查询用户
2. 校验用户状态必须为 `active`，否则 404
3. 根据 `user_nodes` 获取节点集合
4. 通过节点集合获取入站列表
5. 为每个入站生成一个 outbound：
  - `server` = `nodes.public_address`
  - `server_port` = `inbounds.public_port`，若为空则用 `inbounds.listen_port`
  - 其余字段从 `inbounds.settings / tls_settings / transport_settings` 作为客户端模板
  - 用户凭据注入：
    - VLESS/VMess：使用用户 `uuid`
    - Shadowsocks：按模板已有字段（如 `method/password`）注入
6. 按规则返回 JSON 或 Base64(JSON)

## 认证与访问控制
- 订阅为公开链接，仅凭 `user_uuid` 访问
- 无额外 token/signature
- 非 active 用户返回 404

## 错误处理
- `user_uuid` 不存在 → 404
- 用户非 active → 404
- `format` 非 `singbox` → 400
- 生成异常 → 500

## 兼容性与扩展
- UA 映射表内置，后续可扩展 `clash meta / stash / v2ray / surge`
- 后续可加入 `route/dns` 模板、订阅签名或短链

## 支持的协议

### 入站协议（Panel 端验证 + 订阅生成）

| 协议 | 验证器 | 订阅 credential 注入 |
|------|--------|---------------------|
| vless | users 数组非空，uuid/flow 校验 | uuid |
| vmess | users 数组非空，uuid/alterId 校验 | uuid |
| trojan | users 数组非空，password 校验 | password |
| shadowsocks | method 必填，2022 方法多用户校验 | 2022: server_psk:user_key，其他: uuid |
| hysteria2 | users 数组非空，password 校验 | password |
| tuic | users 数组非空，uuid/password 校验 | uuid + password |
| naive | users 数组非空，username/password 校验 | username + password |
| shadowtls | users 数组非空，handshake 必填 | password |
| anytls | users 数组非空，password 校验 | password |
| socks | users 可选，每个 user 校验 username+password | username + password |
| http | 同 socks | username + password |
| mixed | 同 socks | username + password |

### 密钥生成命令（singboxcli）

| 命令 | 说明 |
|------|------|
| `uuid` | 生成 UUID v4 |
| `reality-keypair` | 生成 X25519 REALITY 密钥对 |
| `wg-keypair` | 生成 X25519 WireGuard 密钥对 |
| `vapid-keypair` | 生成 P256 ECDH VAPID 密钥对 |
| `tls-keypair <server_name>` | 生成 TLS 证书（sing-box 库） |
| `rand-base64-<N>` | 生成 N 字节随机 Base64 字符串 |

## 测试重点
- 订阅返回：sing-box UA 返回 JSON，其他 UA 返回 Base64
- `format=singbox` 覆盖 UA
- 用户非 active 404
- 节点分配：只返回用户关联节点的入站
- `public_port` 覆盖逻辑
