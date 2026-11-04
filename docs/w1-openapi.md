# W1 Frozen OpenAPI Operation List

Status: normative for both W1 implementations.
Last updated: 2026-11-04.

This document freezes **paths, methods and operationIds**. Request and response schemas may be extended with optional fields, but a frozen operation may not be renamed, removed, or moved to a different path/method without updating this document.

## Freeze rules

- The five generated specs are `control.yaml`, `ledger.yaml`, `billing.yaml`, `payment.yaml`, `reconcile.yaml`.
- The gateway public API is handwritten and OpenAI-compatible; it is listed here but is not generated from these specs.
- `operationId` is unique inside each spec and is the generated method name.
- Paths are versioned: `/v1` for public and admin, `/internal` for service-to-service, `/webhooks` for PSP callbacks.
- Adding an operation later is allowed with a spec update. Renaming or removing a frozen operation is a breaking change.
- Every operation returns `X-Request-ID`.
- Errors: gateway uses the OpenAI error shape; all other APIs use problem details (see below).

## Authentication

| Scheme | Where | Meaning |
|---|---|---|
| API key | gateway `/v1/*` | `Authorization: Bearer lg_...`; workspace-scoped |
| Session cookie | control/billing/payment/reconcile `/v1/*` | HttpOnly `lg_session` cookie + CSRF for mutating requests |
| Service bearer | `/internal/*` | `Authorization: Bearer <service-token>` from `LG_SERVICE_TOKENS` |
| PSP signature | `/webhooks/*` | Stripe signature header or Alipay signature parameters |

## Gateway public API (handwritten, OpenAI-compatible)

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| POST | `/v1/chat/completions` | `createChatCompletion` | API key | OpenAI-compatible chat completions; streaming and non-streaming; tool calling; vision on OpenAI-routed models |
| GET | `/v1/models` | `listModels` | API key | OpenAI-compatible model list from the model catalog |

Headers:

- Request: `Authorization`, optional `Idempotency-Key`.
- Response: `X-Request-ID`, `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`; `Retry-After` on 429.
- Errors: `{ "error": { "message", "type", "param", "code" } }`.

## control.yaml

Base paths: `/v1` and `/internal`.

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| POST | `/v1/auth/login` | `login` | public | Email/password login; sets session cookie |
| POST | `/v1/auth/logout` | `logout` | session | Clear session |
| GET | `/v1/me` | `getCurrentUser` | session | Current user and memberships |
| GET | `/v1/accounts` | `listAccounts` | session | Accounts visible to the user |
| POST | `/v1/accounts` | `createAccount` | session | Create an account |
| GET | `/v1/accounts/{account_id}` | `getAccount` | session + member | Account detail |
| PATCH | `/v1/accounts/{account_id}` | `updateAccount` | session + owner/admin | Rename or update account |
| GET | `/v1/accounts/{account_id}/workspaces` | `listWorkspaces` | session + member | Workspaces in an account |
| POST | `/v1/accounts/{account_id}/workspaces` | `createWorkspace` | session + owner/admin | Create a workspace |
| GET | `/v1/workspaces/{workspace_id}` | `getWorkspace` | session + member | Workspace detail |
| PATCH | `/v1/workspaces/{workspace_id}` | `updateWorkspace` | session + owner/admin | Rename or update workspace |
| GET | `/v1/workspaces/{workspace_id}/members` | `listMembers` | session + member | List members and roles |
| PUT | `/v1/workspaces/{workspace_id}/members/{user_id}` | `upsertMember` | session + owner/admin | Add member or change role |
| DELETE | `/v1/workspaces/{workspace_id}/members/{user_id}` | `removeMember` | session + owner/admin | Remove member |
| GET | `/v1/workspaces/{workspace_id}/keys` | `listAPIKeys` | session + member | List keys without secrets |
| POST | `/v1/workspaces/{workspace_id}/keys` | `createAPIKey` | session + owner/admin | Create key; secret returned once |
| POST | `/v1/workspaces/{workspace_id}/keys/{key_id}/rotate` | `rotateAPIKey` | session + owner/admin | Rotate key; new secret returned once |
| DELETE | `/v1/workspaces/{workspace_id}/keys/{key_id}` | `revokeAPIKey` | session + owner/admin | Revoke key |
| GET | `/v1/workspaces/{workspace_id}/audit` | `listAuditLog` | session + owner/admin | Audit events |
| POST | `/internal/apikeys/validate` | `validateAPIKey` | service bearer | Validate an API key on gateway cache miss |

## ledger.yaml

Base paths: `/internal` and `/v1`.

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| POST | `/internal/reservations` | `createReservation` | service bearer | Reserve funds for a logical request |
| POST | `/internal/reservations/{reservation_id}/release` | `releaseReservation` | service bearer | Release a hold |
| POST | `/internal/settlements` | `createSettlement` | service bearer | Rate usage and settle a request |
| GET | `/internal/workspaces/{workspace_id}/balance` | `getInternalBalance` | service bearer | Balance for the admin BFF |
| POST | `/internal/adjustments` | `createAdjustment` | service bearer | Post an adjustment from reconciliation |
| GET | `/v1/workspaces/{workspace_id}/balance` | `getBalance` | session | Balance, held, available |
| GET | `/v1/workspaces/{workspace_id}/entries` | `listLedgerEntries` | session | Paginated ledger entries |
| GET | `/v1/workspaces/{workspace_id}/usage` | `listUsage` | session | Paginated usage attempts and settlements |

## billing.yaml

Base paths: `/v1`, `/v1/admin`, `/internal`.

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| GET | `/v1/plans` | `listPlans` | session / public | Active plans |
| GET | `/v1/workspaces/{workspace_id}/subscription` | `getSubscription` | session + member | Current subscription |
| POST | `/v1/workspaces/{workspace_id}/subscription` | `createSubscription` | session + owner/admin | Subscribe or renew a plan |
| DELETE | `/v1/workspaces/{workspace_id}/subscription` | `cancelSubscription` | session + owner/admin | Cancel at period end |
| GET | `/v1/admin/price-catalog` | `listPriceCatalog` | session + admin | Price catalog entries |
| POST | `/v1/admin/price-catalog` | `createPriceCatalogEntry` | session + admin | Add a versioned price entry |
| GET | `/v1/admin/plans` | `listAdminPlans` | session + admin | All plans |
| POST | `/v1/admin/plans` | `createPlan` | session + admin | Create a plan |
| PATCH | `/v1/admin/plans/{plan_id}` | `updatePlan` | session + admin | Update a plan |
| GET | `/internal/price-snapshots/latest` | `getLatestPriceSnapshot` | service bearer | Fallback lookup if a price event was missed |

## payment.yaml

Base paths: `/v1`, `/v1/admin`, `/webhooks`, `/internal`.

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| POST | `/v1/workspaces/{workspace_id}/topups` | `createTopup` | session + owner/admin | Create a Stripe or Alipay checkout |
| GET | `/v1/workspaces/{workspace_id}/topups` | `listTopups` | session + member | Paginated top-ups |
| GET | `/v1/workspaces/{workspace_id}/topups/{payment_intent_id}` | `getTopup` | session + member | Top-up detail |
| GET | `/v1/workspaces/{workspace_id}/payments` | `listPayments` | session + member | Payment history |
| POST | `/v1/admin/refunds` | `createRefund` | session + admin | Issue a refund with a reason code |
| POST | `/webhooks/stripe` | `stripeWebhook` | PSP signature | Stripe webhook |
| POST | `/webhooks/alipay` | `alipayWebhook` | PSP signature | Alipay async notification |
| GET | `/internal/fx-rates` | `getFXRate` | service bearer | Rate for base/quote at a time |
| GET | `/internal/psp-settlements` | `listPSPSettlements` | service bearer | PSP settlement statements for reconciliation |

## reconcile.yaml

Base paths: `/v1/admin` and `/internal`.

| Method | Path | operationId | Auth | Purpose |
|---|---|---|---|---|
| GET | `/v1/admin/reconciliation/runs` | `listReconciliationRuns` | session + admin | Paginated runs |
| POST | `/v1/admin/reconciliation/runs` | `triggerReconciliation` | session + admin | Trigger a run by kind and period |
| GET | `/v1/admin/reconciliation/runs/{run_id}` | `getReconciliationRun` | session + admin | Run detail |
| GET | `/v1/admin/reconciliation/items` | `listReconciliationItems` | session + admin | Filter by status and kind |
| POST | `/v1/admin/reconciliation/items/{item_id}/resolve` | `resolveReconciliationItem` | session + admin | Mark resolved or ignored |
| POST | `/v1/admin/reconciliation/items/{item_id}/adjustments` | `createReconciliationAdjustment` | session + admin | Post a compensating adjustment |
| GET | `/v1/admin/reconciliation/items/{item_id}/adjustments` | `listReconciliationAdjustments` | session + admin | Adjustments for an item |
| POST | `/internal/reconciliation/run` | `runReconciliationInternal` | service bearer | Daemon/test trigger |

## Common conventions

- IDs are UUID strings.
- Timestamps are RFC3339 UTC strings.
- Money fields are integer micro-USD and end in `_micro_usd`.
- Pagination uses `limit` (1–100, default 50) and `cursor`; responses contain `items` and `next_cursor`.
- Mutating admin/control requests require the CSRF token from the session.
- `Idempotency-Key` is supported on `createChatCompletion`; other operations use their natural keys (request ID, payment event ID, subscription ID).
- Error shape for internal/admin APIs:

```json
{
  "type": "https://ledgergate.dev/errors/insufficient_balance",
  "title": "Insufficient balance",
  "status": 402,
  "detail": "The workspace has no available balance for this reservation.",
  "code": "insufficient_balance",
  "request_id": "019..."
}
```

- Error codes are `snake_case`: `invalid_request`, `unauthorized`, `forbidden`, `not_found`, `insufficient_balance`, `rate_limited`, `provider_error`, `internal_error`.
- No operation returns a stack trace or secret.

## OpenAPI file mapping

| File | Service | Generated |
|---|---|---|
| `api/openapi/control.yaml` | control | yes |
| `api/openapi/ledger.yaml` | ledger | yes |
| `api/openapi/billing.yaml` | billing | yes |
| `api/openapi/payment.yaml` | payment | yes |
| `api/openapi/reconcile.yaml` | reconcile | yes |
| gateway public API | gateway | no; handwritten OpenAI-compatible surface |

## Validation

- Every operation in this document appears in the matching OpenAPI file with the same method, path and operationId.
- No frozen operation is missing or renamed.
- `make generate` succeeds from the five specs and leaves no diff on a second run.
- The generated server interfaces compile for all five services.
- W1 handlers may return `501 Not Implemented` for business operations; health/readiness/metrics are separate from these specs.
