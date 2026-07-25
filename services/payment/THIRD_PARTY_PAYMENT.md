# Third-party payment integration

## Account configuration

Use `t_pay_platform.platform_code` as the adapter registry key. Production
credentials should be referenced by `t_tenant_pay_account.credential_ref`; do
not place plaintext keys in `ext_config` or logs.

## Adapter lifecycle

Each provider implementation must implement `internal/provider.Adapter` and be
registered in `internal/svc.NewServiceContext`. The platform-specific adapter
owns request signing, response/status mapping, notification verification and
the exact notification acknowledgement body.

## Recharge flow

1. Create the local recharge order.
2. Commit the local transaction.
3. Call `Adapter.CreatePayment`.
4. Persist the third-party order ID, pay URL/QR content and request log.
5. A verified asynchronous notification marks the payment successful and
   inserts `PAYMENT_RECHARGE_CREDIT` into `t_pay_outbox` in the same database
   transaction.
6. The outbox processor credits Asset RPC with the local order number as its
   idempotency key.

## Notification gateway

The public HTTP gateway must preserve the raw request body, headers and query
parameters and pass them to `Adapter.ParseNotify`. It must not use the normal
JSON response wrapper: return `Adapter.NotifyResponse` exactly as required by
the provider.

Recommended routes:

```text
POST /payment/notify/:platformCode/:accountCode
POST /payment/payout-notify/:platformCode/:accountCode
```

Before accepting a successful notification, validate the signature, merchant
account, local order number, currency and paid amount. Use
`platform_id + notify_id` for callback idempotency.

## Payout flow

Approval moves a withdrawal to `PAYING` and keeps the user's asset frozen.
Only a confirmed payout success may deduct the frozen asset and mark the order
successful. A confirmed payout failure unfreezes the asset. Unknown results
remain frozen and must be queried again.
