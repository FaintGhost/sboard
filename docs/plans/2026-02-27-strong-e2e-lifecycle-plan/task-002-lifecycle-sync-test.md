# Task 002 Test: 生命周期资源编排与手动同步（RED）

**depends-on**: 无

## Description

新增强生命周期主链路的失败测试，覆盖 bootstrap、登录、创建 group/user/node/inbound、绑定关系、手动同步与同步结果验证。该任务先构建失败断言，不实现逻辑补全。

## Execution Context

**Task Number**: 002 of 010  
**Phase**: Core Integration (RED)  
**Prerequisites**: 无

## BDD Scenario

```gherkin
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
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（配置同步与验证）

## Files to Modify/Create

- Create: `e2e/tests/e2e/lifecycle-strong.spec.ts`
- Modify: `e2e/tests/fixtures/api.fixture.ts`
- Modify: `e2e/tests/fixtures/auth.fixture.ts`

## Steps

### Step 1: 构建主链路 RED 用例
- 在新 spec 中声明生命周期场景与断言目标。
- 断言覆盖 group/user/node/inbound 关联完整性与手动同步成功条件。

### Step 2: 声明所需 fixture 能力
- 在 fixture 层声明本场景需要但当前不足的 API 能力（如分组成员替换、同步任务状态查询、可重置的 inbounds 统计读取）。
- 保证 RED 失败为业务断言失败，而非类型或导入失败。

### Step 3: 记录失败基线
- 执行单测并记录失败原因，作为 GREEN 阶段验收输入。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts
```

## Success Criteria

- 生命周期强链路测试文件已创建。
- 测试以预期业务断言失败，失败原因可定位到“能力尚未实现”。
