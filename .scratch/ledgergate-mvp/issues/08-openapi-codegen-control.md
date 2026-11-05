# 08: OpenAPI codegen pipeline and control API stubs

**What to build:** An OpenAPI-first pipeline generates server interfaces from the frozen contracts. The control API surface compiles from the generated server interface, exposes the frozen paths and operation IDs, and needs no business logic for W1.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `make generate` is reproducible and leaves no diff on a second run.
- [ ] The control spec contains exactly the frozen paths, methods and operation IDs.
- [ ] The generated server interface compiles.
- [ ] A contract test asserts the frozen operations.
- [ ] Handlers may return 501; no business logic is required.
