# 11: Billing and reconcile API stubs

**What to build:** Billing and reconcile API surfaces compile from the frozen contracts. Billing exposes plans, subscriptions, price catalog and the latest price snapshot; reconcile exposes runs, items and adjustments. No pricing or reconciliation logic is implemented in W1.

**Blocked by:** 08

**Status:** ready-for-agent

- [ ] Billing and reconcile specs contain exactly the frozen paths, methods and operation IDs.
- [ ] Both generated server interfaces compile.
- [ ] Contract tests assert the frozen operations.
- [ ] No pricing or reconciliation logic is implemented yet.
