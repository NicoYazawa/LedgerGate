# 10: Payment API stubs

**What to build:** The payment API surface compiles from the frozen contract and exposes top-ups, payment history, admin refunds, Stripe and Alipay webhooks, FX lookup and PSP settlement listing. No PSP integration logic is implemented in W1.

**Blocked by:** 08

**Status:** ready-for-agent

- [ ] The payment spec contains exactly the frozen paths, methods and operation IDs.
- [ ] The generated server interface compiles.
- [ ] Webhook routes are present and distinguishable.
- [ ] A contract test asserts the frozen operations.
