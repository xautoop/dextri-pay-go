# Dextri Pay Go SDK

[English](README.md) | 简体中文

Dextri Pay Partner API 的官方 Go SDK，提供通道查询、Hosted Checkout、用户余额、App 自定义兑换、操作记录和 Webhook 验签能力。

## 当前状态

本仓库目前为预发布状态，尚未发布版本标签，首个计划版本为 `v0.1.0`。项目要求 Go 1.25.4 或更高版本，`v1.0.0` 发布前公开 API 仍可能调整。

SDK 可独立安装和使用，不依赖 Dextri 源码目录、数据库、链节点、钱包库或任何私有服务包。

## 安装

正式发布后的版本可通过以下方式安装：

```bash
go get github.com/xautoop/dextri-pay-go@latest
```

## 申请凭证

请通过 [veraxon.xyz](https://veraxon.xyz) 申请 Dextri Pay 接入，凭证将在管理员审核通过后签发。申请时需要提供：

- App 名称和负责人；
- Sandbox 或 Production 环境；
- 允许的回跳域名；
- 需要开通的充值、提现和兑换能力；
- 预计交易额度；
- Webhook 地址，以及可选的出口 IP 白名单。

审核通过后会获得 API Base URL，以及包含 `app_id`、`key_id` 和 `app_secret` 的 App 凭证。Secret 只展示一次，必须保存在服务端密钥管理系统中，不能暴露给浏览器、移动 App、日志或源码仓库。

Sandbox 和 Production 的凭证完全隔离，不能混用。

## 创建客户端

```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/xautoop/dextri-pay-go/client"
)

func newPayClient() *client.Client {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	pay, err := client.New(client.Config{
		BaseURL: os.Getenv("DEXTRI_PAY_BASE_URL"),
		Credentials: client.Credentials{
			AppID:  os.Getenv("DEXTRI_PAY_APP_ID"),
			KeyID:  os.Getenv("DEXTRI_PAY_KEY_ID"),
			Secret: os.Getenv("DEXTRI_PAY_APP_SECRET"),
		},
		UserAgent: "merchant-service/1.0.0",
	}, client.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}
	return pay
}
```

生产地址必须使用 HTTPS。本地开发仅允许通过 `client.WithAllowInsecureLoopbackHTTP()` 显式开启回环地址 HTTP。

## 创建充值 Checkout

所有写操作都必须携带幂等键。每个业务操作应生成并持久化一个幂等键；重试同一业务操作时必须复用原键。

```go
package main

import (
	"context"

	"github.com/xautoop/dextri-pay-go/checkout"
	"github.com/xautoop/dextri-pay-go/client"
	"github.com/xautoop/dextri-pay-go/money"
)

func createDeposit(ctx context.Context, pay *client.Client) (string, error) {
	session, _, err := pay.Checkout.CreateDeposit(ctx, checkout.CreateDepositRequest{
		ExternalUserID:    "user_1001",
		ClientReferenceID: "deposit_20260813_001",
		SourceAsset:       "USDT",
		TargetAsset:       "USDC",
		Amount:            money.Decimal("100.00"),
		ReturnURL:         "https://merchant.example/pay/result",
	}, client.WithIdempotencyKey("deposit_20260813_001"))
	if err != nil {
		return "", err
	}
	return session.CheckoutURL, nil
}
```

App 可以直接展示 `CheckoutURL`，也可以将 `QRPayload` 渲染成二维码。钱包连接和用户授权在 Hosted Checkout 中完成，SDK 不接触用户私钥。

同一个客户端还提供：

- `pay.Channels.List`：查询当前 App 已授权且可用的通道；
- `pay.Checkout.CreateWithdrawal`、`CreateConversion` 和 `CreateDepositAndConvert`；
- `pay.Users.CreateBindingSession` 和 `GetBalances`；
- `pay.Conversions.ListMarkets`、`GetMarket`、`UpdatePrice` 和 `CreateQuote`；
- `pay.Operations.Get` 和 `List`。

资产、市场、denom 和精度均由 API 根据链上及 Admin 配置返回。SDK 不硬编码币种列表，也不使用浮点数做账务计算。所有金额使用 `money.Decimal`，并以 JSON 字符串传输。

## 错误处理

```go
var apiErr *api.APIError
if errors.As(err, &apiErr) {
	log.Printf("code=%s request_id=%s", apiErr.Code, apiErr.RequestID)
}

if api.IsErrorCode(err, "IDEMPOTENCY_CONFLICT") {
	// 同一幂等键不能对应不同请求内容。
}
```

SDK 只会自动重试安全请求，或者携带幂等键的写请求，并遵守 `Retry-After`。调用方必须设置请求截止时间或 `http.Client` 超时。

## Webhook 验签

必须先完成签名校验，再处理事件；业务系统还需要使用 `delivery.Event.ID` 做消费去重。

```go
delivery, err := webhook.Verify(webhookSecret, request.Header, body)
if err != nil {
	// 拒绝请求。
}

event := delivery.Event
```

Webhook Secret 与 App API Secret 相互独立。

## 包说明

- `client`：认证客户端和各业务 API；
- `api`：响应元数据、API 错误和通用 JSON 类型；
- `channels`、`checkout`、`conversion`、`operation`、`users`：出入参类型；
- `money`：字符串形式的十进制金额；
- `webhook`：Webhook 验签和事件类型。

包边界与依赖方向见 [SDK 架构](docs/architecture.zh-CN.md)。

## 仓库开发

无需任何同级仓库即可执行完整检查：

```bash
make check
```

检查内容包括格式、Module 整洁、单元测试、Race Detector、`go vet`、Staticcheck 和 `git diff --check`。

## 许可证

本项目采用 [Apache License 2.0](LICENSE)。
