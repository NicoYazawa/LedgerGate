# W1 Exact Database Schema

Status: normative for both W1 implementations.
Last updated: 2026-11-04.

Both implementations must create the same tables, columns, types, constraints and indexes. `0001_init.sql` for each database is the baseline; W2+ migrations may add columns or tables, but must not rename or retype W1 columns.

## Conventions

- PostgreSQL 18.
- IDs are application-generated UUIDv7 values of type `uuid`; no database defaults.
- Money is `bigint` micro-USD. FX rates are `numeric(20,10)`.
- Timestamps are `timestamptz` in UTC; mutable tables carry `created_at` and `updated_at`.
- Enumerations are `text` plus `CHECK` constraints. Do not use PostgreSQL enum types in W1.
- JSON is `jsonb`. Event payloads are `jsonb`.
- Each service has its own database. No cross-database foreign keys, joins or transactions.
- Goose owns `schema_migrations`; do not create it manually.
- Table names and column names are `snake_case`.
- Index names are `<table>_<purpose>_idx` or `<table>_<purpose>_uq`.

## Common tables

### `outbox`

Append to every database that publishes events: `control`, `ledger`, `billing`, `payment`, `reconcile`. The gateway uses `settlement_outbox` instead, because its only event is a settlement intent.

```sql
CREATE TABLE outbox (
  id uuid PRIMARY KEY,
  event_type text NOT NULL,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  payload jsonb NOT NULL,
  occurred_at timestamptz NOT NULL,
  published_at timestamptz,
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX outbox_unpublished_idx ON outbox (occurred_at) WHERE published_at IS NULL;
```

### `processed_events`

Append to every database that consumes events: `gateway`, `ledger`, `billing`, `reconcile`.

```sql
CREATE TABLE processed_events (
  consumer text NOT NULL,
  event_id uuid NOT NULL,
  processed_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (consumer, event_id)
);
```

## gateway

Tables: `request_log`, `settlement_outbox`, `processed_events`.

```sql
CREATE TABLE request_log (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL,
  api_key_id uuid NOT NULL,
  model_alias text NOT NULL,
  endpoint text NOT NULL,
  stream boolean NOT NULL DEFAULT false,
  status text NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed', 'aborted')),
  http_status integer,
  error_code text,
  idempotency_key text,
  request_hash text,
  attempt_count integer NOT NULL DEFAULT 0,
  started_at timestamptz NOT NULL,
  finished_at timestamptz,
  latency_ms integer,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX request_log_workspace_started_idx ON request_log (workspace_id, started_at DESC);
CREATE INDEX request_log_key_started_idx ON request_log (api_key_id, started_at DESC);
CREATE UNIQUE INDEX request_log_idempotency_uq
  ON request_log (workspace_id, api_key_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE TABLE settlement_outbox (
  id uuid PRIMARY KEY,
  request_id uuid NOT NULL,
  workspace_id uuid NOT NULL,
  reservation_id uuid,
  payload jsonb NOT NULL,
  status text NOT NULL CHECK (status IN ('pending', 'publishing', 'published', 'failed')),
  attempts integer NOT NULL DEFAULT 0,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX settlement_outbox_request_uq ON settlement_outbox (request_id);
CREATE INDEX settlement_outbox_pending_idx ON settlement_outbox (status, next_attempt_at);

-- plus processed_events from the common section
```

Notes:

- One `request_log` row per logical request.
- One `settlement_outbox` row per logical request; it is the gateway's outbox.
- `request_hash` is the hash of the canonical request body used for idempotency comparison.
- `payload` carries the usage attempts, usage source, reservation ID and settlement metadata.

## control

Tables: `accounts`, `users`, `workspaces`, `workspace_members`, `api_keys`, `sessions`, `audit_log`, `outbox`.

```sql
CREATE TABLE accounts (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
  id uuid PRIMARY KEY,
  email text NOT NULL,
  password_hash text NOT NULL,
  display_name text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_uq ON users (lower(email));

CREATE TABLE workspaces (
  id uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES accounts(id),
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workspaces_account_idx ON workspaces (account_id);

CREATE TABLE workspace_members (
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE api_keys (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name text NOT NULL,
  key_prefix text NOT NULL,
  key_hash text NOT NULL,
  last_four text NOT NULL,
  model_allowlist jsonb,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
  expires_at timestamptz,
  last_used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX api_keys_hash_uq ON api_keys (key_hash);
CREATE INDEX api_keys_workspace_status_idx ON api_keys (workspace_id, status);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash text NOT NULL,
  csrf_token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX sessions_token_uq ON sessions (token_hash);
CREATE INDEX sessions_user_expires_idx ON sessions (user_id, expires_at);

CREATE TABLE audit_log (
  id uuid PRIMARY KEY,
  actor_type text NOT NULL CHECK (actor_type IN ('user', 'admin', 'system')),
  actor_id uuid,
  action text NOT NULL,
  target_type text NOT NULL,
  target_id uuid,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  ip inet,
  user_agent text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_created_idx ON audit_log (created_at DESC);
CREATE INDEX audit_log_target_idx ON audit_log (target_type, target_id);

-- plus outbox from the common section
```

Notes:

- `users.email` is unique case-insensitively via the `lower(email)` index.
- `api_keys.model_allowlist` is a JSON array of model aliases; `null` means all models.
- `api_keys.key_prefix` is the visible prefix; `key_hash` is the HMAC-SHA256 hex digest.
- `control` publishes `apikey.created` and `apikey.revoked` through its outbox.

## ledger

Tables: `ledger_accounts`, `wallet_balances`, `ledger_entries`, `reservations`, `settlements`, `usage_attempts`, `price_snapshots`, `credit_lots`, `outbox`, `processed_events`.

```sql
CREATE TABLE ledger_accounts (
  id uuid PRIMARY KEY,
  code text NOT NULL,
  type text NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
  workspace_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ledger_accounts_code_uq ON ledger_accounts (code);
CREATE UNIQUE INDEX ledger_accounts_workspace_uq
  ON ledger_accounts (workspace_id)
  WHERE workspace_id IS NOT NULL;

CREATE TABLE wallet_balances (
  workspace_id uuid PRIMARY KEY,
  ledger_account_id uuid NOT NULL REFERENCES ledger_accounts(id),
  balance_micro_usd bigint NOT NULL DEFAULT 0,
  held_micro_usd bigint NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (held_micro_usd >= 0)
);

CREATE UNIQUE INDEX wallet_balances_account_uq ON wallet_balances (ledger_account_id);

CREATE TABLE ledger_entries (
  id uuid PRIMARY KEY,
  transaction_id uuid NOT NULL,
  account_id uuid NOT NULL REFERENCES ledger_accounts(id),
  workspace_id uuid,
  direction text NOT NULL CHECK (direction IN ('debit', 'credit')),
  amount_micro_usd bigint NOT NULL CHECK (amount_micro_usd > 0),
  currency text NOT NULL DEFAULT 'USD' CHECK (currency = 'USD'),
  reference_type text NOT NULL CHECK (reference_type IN (
    'settlement', 'topup', 'refund', 'chargeback', 'adjustment', 'subscription_grant'
  )),
  reference_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ledger_entries_account_created_idx ON ledger_entries (account_id, created_at DESC);
CREATE INDEX ledger_entries_transaction_idx ON ledger_entries (transaction_id);
CREATE INDEX ledger_entries_reference_idx ON ledger_entries (reference_type, reference_id);
CREATE UNIQUE INDEX ledger_entries_transaction_account_direction_uq
  ON ledger_entries (transaction_id, account_id, direction);

CREATE TABLE reservations (
  id uuid PRIMARY KEY,
  request_id uuid NOT NULL,
  workspace_id uuid NOT NULL,
  reserved_micro_usd bigint NOT NULL CHECK (reserved_micro_usd > 0),
  status text NOT NULL CHECK (status IN ('held', 'settled', 'released', 'expired')),
  expires_at timestamptz NOT NULL,
  settled_at timestamptz,
  released_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX reservations_request_uq ON reservations (request_id);
CREATE INDEX reservations_expiry_idx ON reservations (status, expires_at);

CREATE TABLE price_snapshots (
  id uuid PRIMARY KEY,
  provider text NOT NULL,
  model text NOT NULL,
  input_rate_micro_usd_per_million bigint NOT NULL,
  output_rate_micro_usd_per_million bigint NOT NULL,
  cache_read_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  cache_write_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  currency text NOT NULL DEFAULT 'USD' CHECK (currency = 'USD'),
  effective_from timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX price_snapshots_lookup_idx
  ON price_snapshots (provider, model, effective_from DESC);

CREATE TABLE settlements (
  id uuid PRIMARY KEY,
  request_id uuid NOT NULL,
  reservation_id uuid NOT NULL REFERENCES reservations(id),
  workspace_id uuid NOT NULL,
  charged_micro_usd bigint NOT NULL DEFAULT 0,
  provider_cost_micro_usd bigint NOT NULL DEFAULT 0,
  currency text NOT NULL DEFAULT 'USD' CHECK (currency = 'USD'),
  price_snapshot_id uuid NOT NULL REFERENCES price_snapshots(id),
  usage_source text NOT NULL CHECK (usage_source IN ('provider', 'estimated', 'mixed', 'none')),
  status text NOT NULL CHECK (status IN ('settled', 'failed')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX settlements_request_uq ON settlements (request_id);
CREATE INDEX settlements_workspace_created_idx ON settlements (workspace_id, created_at DESC);

CREATE TABLE usage_attempts (
  id uuid PRIMARY KEY,
  request_id uuid NOT NULL,
  settlement_id uuid REFERENCES settlements(id),
  attempt_no integer NOT NULL CHECK (attempt_no > 0),
  provider text NOT NULL,
  model text NOT NULL,
  input_tokens bigint NOT NULL DEFAULT 0,
  output_tokens bigint NOT NULL DEFAULT 0,
  cache_read_tokens bigint NOT NULL DEFAULT 0,
  cache_write_tokens bigint NOT NULL DEFAULT 0,
  reasoning_tokens bigint NOT NULL DEFAULT 0,
  usage_source text NOT NULL CHECK (usage_source IN ('provider', 'estimated', 'missing')),
  provider_cost_micro_usd bigint NOT NULL DEFAULT 0,
  sell_cost_micro_usd bigint NOT NULL DEFAULT 0,
  status text NOT NULL CHECK (status IN ('succeeded', 'failed', 'aborted')),
  error_code text,
  started_at timestamptz NOT NULL,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX usage_attempts_request_attempt_uq ON usage_attempts (request_id, attempt_no);
CREATE INDEX usage_attempts_settlement_idx ON usage_attempts (settlement_id);
CREATE INDEX usage_attempts_provider_model_idx
  ON usage_attempts (provider, model, created_at DESC);

CREATE TABLE credit_lots (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL,
  source_type text NOT NULL CHECK (source_type IN ('subscription_grant', 'promo')),
  source_id uuid NOT NULL,
  original_micro_usd bigint NOT NULL CHECK (original_micro_usd > 0),
  remaining_micro_usd bigint NOT NULL CHECK (remaining_micro_usd >= 0),
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX credit_lots_source_uq ON credit_lots (source_type, source_id);
CREATE INDEX credit_lots_workspace_expiry_idx ON credit_lots (workspace_id, expires_at);

-- plus outbox and processed_events from the common section
```

Notes:

- `wallet_balances.balance_micro_usd` may go negative only through chargebacks, late settlement, or admin adjustment.
- `available = balance_micro_usd - held_micro_usd`. The application enforces `available >= reserved` atomically.
- `ledger_entries` is append-only. Corrections are new entries with a new `transaction_id`.
- `settlements.request_id` is unique: one settlement per logical request.
- `usage_attempts` has one row per upstream attempt. Every attempt is recorded even when the user is charged once.
- `price_snapshots` in the ledger is a read-only copy of the billing snapshot with the same ID.
- `credit_lots` is consumed earliest-expiring first before wallet balance.

## billing

Tables: `price_catalog`, `price_snapshots`, `plans`, `subscriptions`, `grant_policies`, `outbox`, `processed_events`.

```sql
CREATE TABLE price_catalog (
  id uuid PRIMARY KEY,
  provider text NOT NULL,
  model text NOT NULL,
  currency text NOT NULL CHECK (currency IN ('USD', 'CNY')),
  cost_input_rate_micro_usd_per_million bigint NOT NULL,
  cost_output_rate_micro_usd_per_million bigint NOT NULL,
  cost_cache_read_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  cost_cache_write_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  sell_input_rate_micro_usd_per_million bigint NOT NULL,
  sell_output_rate_micro_usd_per_million bigint NOT NULL,
  sell_cache_read_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  sell_cache_write_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  markup_bps integer NOT NULL DEFAULT 0,
  effective_from timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX price_catalog_unique_idx
  ON price_catalog (provider, model, currency, effective_from);

CREATE TABLE price_snapshots (
  id uuid PRIMARY KEY,
  catalog_id uuid NOT NULL REFERENCES price_catalog(id),
  provider text NOT NULL,
  model text NOT NULL,
  currency text NOT NULL CHECK (currency IN ('USD', 'CNY')),
  cost_input_rate_micro_usd_per_million bigint NOT NULL,
  cost_output_rate_micro_usd_per_million bigint NOT NULL,
  cost_cache_read_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  cost_cache_write_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  sell_input_rate_micro_usd_per_million bigint NOT NULL,
  sell_output_rate_micro_usd_per_million bigint NOT NULL,
  sell_cache_read_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  sell_cache_write_rate_micro_usd_per_million bigint NOT NULL DEFAULT 0,
  effective_from timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX price_snapshots_lookup_idx
  ON price_snapshots (provider, model, currency, effective_from DESC);

CREATE TABLE plans (
  id uuid PRIMARY KEY,
  code text NOT NULL,
  name text NOT NULL,
  monthly_price_micro_usd bigint NOT NULL CHECK (monthly_price_micro_usd >= 0),
  included_credit_micro_usd bigint NOT NULL CHECK (included_credit_micro_usd >= 0),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX plans_code_uq ON plans (code);

CREATE TABLE subscriptions (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL,
  plan_id uuid NOT NULL REFERENCES plans(id),
  status text NOT NULL CHECK (status IN ('active', 'canceled', 'expired')),
  current_period_start timestamptz NOT NULL,
  current_period_end timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX subscriptions_active_workspace_uq
  ON subscriptions (workspace_id)
  WHERE status = 'active';
CREATE INDEX subscriptions_workspace_idx ON subscriptions (workspace_id, created_at DESC);

CREATE TABLE grant_policies (
  id uuid PRIMARY KEY,
  plan_id uuid NOT NULL REFERENCES plans(id),
  amount_micro_usd bigint NOT NULL CHECK (amount_micro_usd > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX grant_policies_plan_idx ON grant_policies (plan_id);

-- plus outbox and processed_events from the common section
```

Notes:

- CNY rows are presentment price rows; the ledger only stores USD snapshots. The billing service publishes USD snapshots to the ledger.
- `markup_bps` is basis points; `12000` means 1.2x.
- `grant_policies` is the plan-level grant template; the actual grant is a ledger `credit_lot`.

## payment

Tables: `payment_intents`, `topups`, `refunds`, `webhook_events`, `psp_settlements`, `fx_rates`, `outbox`.

```sql
CREATE TABLE payment_intents (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL,
  provider text NOT NULL CHECK (provider IN ('stripe', 'alipay')),
  provider_ref text,
  status text NOT NULL CHECK (status IN ('created', 'pending', 'succeeded', 'failed', 'refunded')),
  presentment_currency text NOT NULL CHECK (presentment_currency IN ('USD', 'CNY')),
  presentment_amount_minor bigint NOT NULL CHECK (presentment_amount_minor > 0),
  settled_usd_micro_usd bigint,
  fx_rate numeric(20,10),
  psp_fee_micro_usd bigint NOT NULL DEFAULT 0,
  failure_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  succeeded_at timestamptz
);

CREATE UNIQUE INDEX payment_intents_provider_ref_uq
  ON payment_intents (provider, provider_ref)
  WHERE provider_ref IS NOT NULL;
CREATE INDEX payment_intents_workspace_idx
  ON payment_intents (workspace_id, created_at DESC);

CREATE TABLE topups (
  id uuid PRIMARY KEY,
  payment_intent_id uuid NOT NULL REFERENCES payment_intents(id),
  workspace_id uuid NOT NULL,
  credited_micro_usd bigint NOT NULL CHECK (credited_micro_usd > 0),
  status text NOT NULL CHECK (status IN ('pending', 'credited', 'reversed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX topups_intent_uq ON topups (payment_intent_id);

CREATE TABLE refunds (
  id uuid PRIMARY KEY,
  payment_intent_id uuid NOT NULL REFERENCES payment_intents(id),
  amount_presentment_minor bigint NOT NULL CHECK (amount_presentment_minor > 0),
  refunded_usd_micro_usd bigint NOT NULL CHECK (refunded_usd_micro_usd > 0),
  fx_rate numeric(20,10),
  status text NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed')),
  reason_code text NOT NULL,
  actor_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refunds_intent_idx ON refunds (payment_intent_id);

CREATE TABLE webhook_events (
  id uuid PRIMARY KEY,
  provider text NOT NULL CHECK (provider IN ('stripe', 'alipay')),
  event_id text NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  status text NOT NULL CHECK (status IN ('received', 'processed', 'failed')),
  processed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX webhook_events_provider_event_uq
  ON webhook_events (provider, event_id);

CREATE TABLE psp_settlements (
  id uuid PRIMARY KEY,
  provider text NOT NULL CHECK (provider IN ('stripe', 'alipay')),
  settlement_ref text NOT NULL,
  currency text NOT NULL CHECK (currency IN ('USD', 'CNY')),
  amount_minor bigint NOT NULL,
  fee_minor bigint NOT NULL DEFAULT 0,
  settled_at timestamptz NOT NULL,
  raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX psp_settlements_ref_uq
  ON psp_settlements (provider, settlement_ref);

CREATE TABLE fx_rates (
  id uuid PRIMARY KEY,
  base_currency text NOT NULL CHECK (base_currency IN ('USD', 'CNY')),
  quote_currency text NOT NULL CHECK (quote_currency IN ('USD', 'CNY')),
  rate numeric(20,10) NOT NULL CHECK (rate > 0),
  source text NOT NULL CHECK (source IN ('psp', 'manual')),
  effective_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX fx_rates_lookup_idx
  ON fx_rates (base_currency, quote_currency, effective_at DESC);

-- plus outbox from the common section
```

Notes:

- `presentment_amount_minor` is cents for USD and fen for CNY.
- `settled_usd_micro_usd` and `fx_rate` are filled when the PSP confirms success.
- `webhook_events` is the PSP webhook idempotency table; `(provider, event_id)` is unique.
- `psp_settlements` is the settlement/payout statement used by reconciliation.

## reconcile

Tables: `reconciliation_runs`, `reconciliation_items`, `adjustments`, `outbox`, `processed_events`.

```sql
CREATE TABLE reconciliation_runs (
  id uuid PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('internal', 'provider', 'psp', 'subscription')),
  period_start timestamptz NOT NULL,
  period_end timestamptz NOT NULL,
  status text NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX reconciliation_runs_kind_period_idx
  ON reconciliation_runs (kind, period_start DESC);

CREATE TABLE reconciliation_items (
  id uuid PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES reconciliation_runs(id),
  kind text NOT NULL CHECK (kind IN ('internal', 'provider', 'psp', 'subscription')),
  reference_type text NOT NULL,
  reference_id uuid,
  expected_micro_usd bigint,
  actual_micro_usd bigint,
  difference_micro_usd bigint,
  currency text NOT NULL CHECK (currency IN ('USD', 'CNY')),
  status text NOT NULL CHECK (status IN ('open', 'resolved', 'ignored')),
  details jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX reconciliation_items_status_idx
  ON reconciliation_items (status, created_at);
CREATE INDEX reconciliation_items_run_idx
  ON reconciliation_items (run_id);

CREATE TABLE adjustments (
  id uuid PRIMARY KEY,
  reconciliation_item_id uuid REFERENCES reconciliation_items(id),
  ledger_transaction_id uuid,
  amount_micro_usd bigint NOT NULL,
  reason text NOT NULL,
  actor_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX adjustments_item_idx ON adjustments (reconciliation_item_id);

-- plus outbox and processed_events from the common section
```

Notes:

- `ledger_transaction_id` is a cross-service reference and intentionally has no foreign key.
- Adjustments create compensating ledger entries through the ledger API; reconciliation never edits ledger rows directly.

## Ownership summary

| Database | Tables |
|---|---|
| `ledgergate_gateway` | `request_log`, `settlement_outbox`, `processed_events` |
| `ledgergate_control` | `accounts`, `users`, `workspaces`, `workspace_members`, `api_keys`, `sessions`, `audit_log`, `outbox` |
| `ledgergate_ledger` | `ledger_accounts`, `wallet_balances`, `ledger_entries`, `reservations`, `settlements`, `usage_attempts`, `price_snapshots`, `credit_lots`, `outbox`, `processed_events` |
| `ledgergate_billing` | `price_catalog`, `price_snapshots`, `plans`, `subscriptions`, `grant_policies`, `outbox`, `processed_events` |
| `ledgergate_payment` | `payment_intents`, `topups`, `refunds`, `webhook_events`, `psp_settlements`, `fx_rates`, `outbox` |
| `ledgergate_reconcile` | `reconciliation_runs`, `reconciliation_items`, `adjustments`, `outbox`, `processed_events` |

## Validation

- Apply each `0001_init.sql` to a fresh PostgreSQL 18 database.
- Run it twice: goose tracks migrations, so the second application must be a no-op.
- Check `\d+ <table>` for every table; column types and nullability must match this document.
- Foreign keys must resolve inside the same database only.
- No cross-database views, foreign keys, or joins.
