# SDK 架构

Dextri Pay Go 使用一个明确的 Client 门面、按能力划分的公开包，以及小而专一的内部基础设施包。

## 依赖方向

```text
商户应用
   |
   v
client.Client 与各能力 Service
        |                         |
        v                         v
公开请求/响应类型              internal/transport
                                  |       |
                                  v       v
                            internal/auth internal/retry
```

`client.Client` 在创建时完成全部能力 Service 的装配。每个 Service 负责串联稳定请求校验、API 路径映射和响应契约，并通过私有 `executor` 接口调用 Transport。架构中刻意不保留第二层 Resource/Provider 转发。

公开领域包只保存 API 值类型和稳定的校验规则，不依赖内部实现。Transport 保持领域无关，不能反向依赖具体业务能力包。

Webhook 与 API Client 相互独立。事件类型、旧载荷兼容和签名验证按文件分离，但继续保持为一个公开 `webhook` 包。

`internal/architecture` 下的架构测试负责持续约束这些边界。
