# 09: Ledger API stubs

**What to build:** The ledger API surface compiles from the frozen OpenAPI contract and exposes the reserve, release, settle, balance, entries, usage and adjustment operations. No money logic is implemented in W1.

**Blocked by:** 08

**Status:** ready-for-agent

- [ ] The ledger spec contains exactly the frozen paths, methods and operation IDs.
- [ ] The generated server interface compiles.
- [ ] A contract test asserts the frozen operations.
- [ ] No settlement or balance logic is implemented yet.
