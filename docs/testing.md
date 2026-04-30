# dtkt-integrations testing

Project-specific testing rules for integration packages. Universe-wide Go
testing conventions live in
[../../dtkt-dev/docs/go/testing.md](../../dtkt-dev/docs/go/testing.md) - read
that first.

This file documents what is unique to `dtkt-integrations`: flows are the
integration test surface, not Go-level harnesses.

## Use this for that

### Flows are the test surface

- **Add a representative flow under [../flows/](../flows/)** when adding a new
  service or RPC. Flows are how integrations are consumed in production, so
  the flow doubles as the reference test.
- **Action `call.method` is fully-qualified** (`ai.v1beta1.AgentService.Run`),
  not bare. The platform routes by service name; a bare method name will not
  match.
- **Each new service or RPC gets at least one flow.** Reviewers will block PRs
  that add a service without an exercising flow.

### Per-package Go tests

- **`task test` per package.** Each `packages/<name>/` runs Go tests via the
  shared `Taskfile.go.yaml` (`task go:test`) wired in by
  [`${TASK_PATH}/Taskfile.go.yaml`](#) per [style.md](style.md).
- **Unit tests live next to the code under `pkg/`**; do not invent a separate
  `tests/` tree. Cross-package tests do not exist - integration packages are
  isolated by design (one `go.mod` per package, no `replace` directives).
- **`CheckAuth` and `Close` need exercising.** A package whose `CheckAuth` is
  never called from a test or flow has not been validated end-to-end.

### CGO packages

- **CGO test runs need the CGO toolchain.** `packages/bigquery/` is the only
  CGO integration today (`CGO_ENABLED=1`, `CXX=clang++`). Tests that touch CGO
  paths must be runnable in CI with the same env; do not gate them only on a
  local toolchain.

## Patterns to add

> Add as the surface grows - canonical flow shape for managed event sources,
> retry/backoff assertion patterns, error-mapping coverage to gRPC `codes.*`.
