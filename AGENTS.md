# AGENTS.md

This file provides guidance for AI coding agents working in this repository.

## Overview

`dtkt-integrations` houses the open-source integrations that extend DataKit to external services (BigQuery, OpenAI, Postmark, etc.) by implementing the Protobuf interfaces defined in [dtkt-sdk](https://github.com/datakit-dev/dtkt-sdk). Each subdirectory under [packages/](packages/) is a **standalone Go service** with its own `go.mod`, `Dockerfile`, `Taskfile.yaml`, and `package.dtkt.yaml` manifest. Integrations register themselves with the platform via gRPC service descriptors in their `main.go` (catalog, replication, ai, email, etc.) and are exercised by [flows/](flows/) - declarative `Flow` YAML that wires connections, inputs, actions, and outputs against those services.

Two ways to run integrations locally:

- **`dtkt intgr dev` per package** (the canonical author workflow). [start-all.sh](start-all.sh) and [stop-all.sh](stop-all.sh) walk every `package.dtkt.yaml` under `packages/` and run `dtkt i start -d` / `dtkt i stop` in each - useful for spinning up the full set against a connected DataKit Cloud Network.
- **Skaffold against a local kind cluster** ([skaffold.yaml](skaffold.yaml)) - orchestrates building images and deploying each integration as a k8s Deployment via the per-package kustomize overlays. This is what `task dev` drives.

This repo is cloned standalone and assumes [dtkt-cli](https://github.com/datakit-dev/dtkt-cli) (`dtkt`) is installed and on `PATH`. See [CONTRIBUTING.md](CONTRIBUTING.md) and [README.md](README.md) for one-time setup.

## Commands

All Go builds set `GOEXPERIMENT=jsonv2`. The root [Taskfile.yaml](Taskfile.yaml) fans out across the `GO_INTEGRATIONS` list (`bigquery`, `command`, `email`, `localblob`, `mailpit`, `metabusiness`, `openai`, `postmark`).

| Goal | Command |
| --- | --- |
| First-time setup (sops/terraform/kustomize/kind preflight + per-package `task setup`) | `task setup` |
| Run the dev loop on kind (skaffold dev, all enabled integrations + infra) | `task dev` |
| Run all integrations once (no watcher) | `task run` |
| Tear down the dev environment (keeps PVs) | `task delete` |
| Tear down everything including PVs | `task delete-all` |
| Code generation across all packages | `task generate` |
| Lint / build / test all packages | `task lint` / `task build` / `task test` |
| `go mod tidy` across all packages | `task tidy` |
| Watch pods in the local namespace | `task watch-pods-local` |
| Single integration dev loop | `(cd packages/<name> && task dev)` (runs `dtkt intgr dev`) |
| Single integration build/lint/test | `(cd packages/<name> && task build|lint|test)` |
| Start every integration locally via the CLI | `./start-all.sh` |
| Stop every running integration | `./stop-all.sh` |
| Build every integration via the CLI | `./hack/scripts/build-all.sh` |

`task dev` accepts skaffold profile overrides via `CLI_ARGS`, e.g. `task dev -- --profile bigquery`. Without overrides it runs `--profile all`.

The [skaffold.yaml](skaffold.yaml) profile list is the source of truth for which integrations are wired into the kind dev loop. Several packages (`command`, `email`, `postmark`, `mailpit`, `localblob`) are currently commented out - they still build and run via `dtkt intgr dev` but are not part of `task dev`.

## Architecture

### Per-package shape

Each `packages/<name>/` is a self-contained Go module that follows this convention (see [packages/bigquery/](packages/bigquery/), [packages/openai/](packages/openai/), [packages/email/](packages/email/) for reference):

- `main.go` - constructs an `integrationsdk.Integration` from the embedded `package.dtkt.yaml`, calls `integrationsdk.RegisterService(...)` for each gRPC service descriptor the integration implements, optionally registers managed action/event services, then `intgr.Serve()`.
- `package.dtkt.yaml` - the integration manifest (`kind: Package`, `apiVersion: v1beta1`). Sets `identity.name`, `version`, `description`, `icon`, and `type: PACKAGE_TYPE_GO`. Embedded into the binary via `//go:embed`.
- `pkg/` - implementation: `instance.go` (the per-instance config / state struct passed to `NewInstance`), per-service files (`email_service.go`, `embedding.go`, `action.go`, `event.go`), and any internal helpers (`pkg/lib/`, transport-specific subpackages like `imap/`, `smtp/`, `pop3/`, codegen sinks like `pkg/oapigen/`). Versioned subpackages (e.g. `pkg/v1beta1/`, `pkg/v1beta2/`) when the package implements versioned proto APIs side by side.
- `Taskfile.yaml` - local tasks. Includes `${TASK_PATH}/Taskfile.go.yaml` (with `GO_BUILD_BINARY: pkg`), `${TASK_PATH}/Taskfile.sops.yaml`, and `${TASK_PATH}/Taskfile.kustomize.yaml` mounted twice - once at `infra/kustomize/overlays/dev` (alias `k:dev`) and once at `infra/kustomize/overlays/prd` (alias `k:prd`). Standard tasks: `setup`, `dev` (just `dtkt intgr dev`), `generate` (delegates to `go:generate`; packages with proto also call `buf generate`), `lint`, `build`, `test`, `vhs-tapes`, `docs`. `${TASK_PATH}` is an environment variable pointing at a shared Taskfile directory provided by the dev environment - it must be set in the shell before invoking these tasks.
- `Dockerfile` + `Dockerfile.dockerignore` - container image for k8s deployment.
- `skaffold.yaml` - single profile named after the package, builds the Dockerfile from the universe root context (`../../../`) and applies `infra/kustomize/overlays/dev`. Sets a port-forward (`bigquery` -> `8200`, etc.).
- `infra/kustomize/{base,overlays/dev,overlays/prd}` - k8s manifests. Secrets are sops-encrypted.
- `buf.yaml` / `buf.gen.yaml` / `proto/` - only when the integration ships its own proto definitions (most do; `openai` uses `oapigen.yaml` for OpenAPI codegen instead).

### Repo-level layout

- [packages/](packages/) - the integrations themselves (one Go module each, separate `go.mod`).
- [flows/](flows/) - `Flow` YAML manifests that exercise integration services end-to-end (e.g. `get_weather.yaml` calls `ai.v1beta1.AgentService.Run`, `bigquery_audit_logs.yaml`, `openai_realtime_chat.yaml`). These are the integration test surface and the reference for how integrations are consumed.
- [infra/](infra/) - repo-wide infra: [infra/terraform/](infra/terraform/) (workspace-aware OpenTofu - `main.tf`, `providers.tf`, `state.tf`, `secrets.tf`, `github.tf`, with `dev.tfvars` / `prd.tfvars` and sops-encrypted `secrets/`), [infra/kustomize/](infra/kustomize/) (shared base + overlays), and [infra/skaffold.yaml](infra/skaffold.yaml) (the `infra` profile pulled in by the root skaffold via `requires`).
- [hack/scripts/](hack/scripts/) - `build-all.sh` / `start-all.sh` / `stop-all.sh` shell helpers that drive `dtkt i ...` over every `package.dtkt.yaml` they find. The repo-root `start-all.sh` / `stop-all.sh` are the same idea, executed from `.`.
- [docs/](docs/) - vhs tape sources for generated terminal recordings.
- [dtkt-integrations.code-workspace](dtkt-integrations.code-workspace) - multi-root VSCode workspace; opens `bigquery`, `metabusiness`, `openai`, and the repo root as separate folders. Sets `GOEXPERIMENT=jsonv2` and rulers at 80/120.

### Where things wire together

`main.go` in each package is the registration point. The patterns to recognize (verbatim from existing packages):

- `integrationsdk.NewFS(<embed.FS>, pkg.NewInstance)` - constructs the integration from the embedded `package.dtkt.yaml` and a per-instance constructor.
- `integrationsdk.RegisterService(intgr, &<svc>.<Service>_ServiceDesc, pkg.New<Service>)` - a plain gRPC service implementation. Registered once per service the integration exposes (e.g. bigquery registers `CatalogService`, `SchemaService`, `TableService`, `QueryService`, `GeoService`).
- `integrationsdk.RegisterManagedActionService(intgr, pkg.NewActionService, pkg.<Actions>()...)` - actions exposed to flows (openai uses this for realtime actions).
- `integrationsdk.RegisterManagedEventService(intgr, pkg.NewEventService, events, source...)` - event streams. Sources include webhook events (`fivetran.WebhookEventSource(intgr)`), audit logs (`lib.AuditLogSource[*pkgv1beta1.Instance]()`), and integration-specific feeds (`pkg.EmailEventSource()`, `pkg.RealtimeSource()`).
- `intgr.Serve()` - blocks; failure is fatal.

`flows/*.yaml` reference these services by their fully-qualified proto name under `connections[].services[]` (`ai.v1beta1.AgentService`, `email.v1beta1.EmailService`, `replication.v1beta1.DestinationService`, …). The flow's `actions[].call.method` then names a specific RPC on that service. A flow is only runnable when an integration registered for that service is reachable on the network.

## Adding a new integration

Mirror an existing package (`packages/email/` is a minimal reference; `packages/bigquery/` is a fuller one):

1. Create `packages/<name>/` with `go.mod` (`module github.com/datakit-dev/dtkt-integrations/<name>`), `main.go`, `pkg/`, `Dockerfile`, `Dockerfile.dockerignore`, `Taskfile.yaml`, `package.dtkt.yaml`, `skaffold.yaml`, and `infra/kustomize/{base,overlays/dev,overlays/prd}`.
2. In `package.dtkt.yaml` set `identity.name`, `description`, `icon`, and `type: PACKAGE_TYPE_GO`.
3. In `main.go`, embed the manifest (`//go:embed package.dtkt.yaml`), construct the integration with `integrationsdk.NewFS(...)`, register each gRPC service the integration implements, then call `intgr.Serve()`.
4. Copy the `Taskfile.yaml` shape from a sibling - it should include `Taskfile.go.yaml` with `GO_BUILD_BINARY: pkg` and expose `setup`/`dev`/`generate`/`lint`/`build`/`test`.
5. Add the package name to `GO_INTEGRATIONS` in the root [Taskfile.yaml](Taskfile.yaml) so it's picked up by `task setup` / `task generate` / `task lint` / `task build` / `task test` / `task tidy`.
6. Add a `requires` entry (and matching `profiles` entry) for the package in [skaffold.yaml](skaffold.yaml) if it should run in the kind dev loop.
7. If the integration ships its own proto, add `buf.yaml` + `buf.gen.yaml` and call `buf generate` from the package's `generate` task (see `packages/email/Taskfile.yaml`).
8. Add a representative flow under [flows/](flows/) that exercises the new service.

### The `Instance` type

The `pkg.Instance` struct (constructed via `pkg.NewInstance(ctx, config *...intgr.Config) (*Instance, error)`) is the per-instance state for an integration. It holds the decoded `Config` proto plus any clients/sub-services it owns (e.g. `smtp`, `template`, `bigquery client`). Two methods are required by the SDK:

- `CheckAuth(ctx, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error)` - auth probe. Stub with `status.Errorf(codes.Unimplemented, ...)` if the integration doesn't expose auth, but most production integrations should implement this.
- `Close() error` - release client connections / pools when the instance is torn down.

The `Instance` type parameter (e.g. `*pkgv1beta1.Instance` in bigquery) is what flows through generic helpers like `lib.AuditLogEvents[*pkgv1beta1.Instance]()` and `v1beta1.InstanceMux[*pkgv1beta1.Instance]` - keep it consistent across `main.go` and the `pkg/` files.

### Flow YAML

Flows under [flows/](flows/) follow a small DSL. A typical flow declares:

- `connections[]` - named connection slots, each typed by one or more proto services it requires.
- `inputs[]` - typed entry points (`string`, with `nullable: false`, etc.).
- `actions[]` - RPC calls. Each action's `call.connection` references a `connections[].id`, `call.method` is a fully-qualified RPC name, and `call.request` is a CEL-like expression block (e.g. `"= \"What is the current weather in ${inputs.location.getValue()}\""`).
- `outputs[]` - surfaced values, typically `= actions.<id>.getValue()`.

Flows are validated against `dtkt.flowsdk.v1beta1.Flow` (the schema is referenced by yaml-language-server comments at the top of each file). When adding a service or RPC to an integration, add or update a flow that exercises it - flows are how the platform proves the integration end-to-end.

## Gotchas

- **`task dev` is skaffold, not the CLI dev loop.** It deploys to the `kind-dtkt-local` cluster in the `dtkt-integrations` namespace. The kind cluster must already exist (`kind create cluster --name dtkt-local`) and docker must be running - the `setup`/`dev` preconditions check this. The CLI-driven workflow (`dtkt intgr dev`, `start-all.sh`) is independent and does not touch kind.
- **Skaffold profile list lags the package list.** Several packages are commented out in [skaffold.yaml](skaffold.yaml). Don't assume `task dev --profile all` covers every integration - check the file.
- **Skaffold build context is the universe root** (`../../../`) and the Dockerfile path is `dtkt-integrations/packages/<name>/Dockerfile`. New packages must follow this pattern; building from the package directory directly will not resolve sibling repos.
- **`package.dtkt.yaml` is embedded into the binary.** Editing it requires a rebuild - `dtkt intgr dev` watches and rebuilds, but plain `go run` won't pick up changes after the first build if the manifest changed.
- **`bigquery/Taskfile.yaml` lint is currently a TODO** (`echo TODO: fix go lint` rather than `task go:lint`). Don't copy this into a new integration; use the form from `packages/openai/Taskfile.yaml`.
- **Secrets are sops-encrypted** under `infra/`. Don't commit decrypted versions. (See [dtkt-dev/docs/sops.md](../dtkt-dev/docs/sops.md).)
- **Terraform is workspace-aware** via `Taskfile.tofu-workspace.yaml` with `dev` and `prd` workspaces (`ALL_WORKSPACES: dev prd`). Set `WORKSPACE=dev` (or `prd`) when running `tf:*` tasks that target a single environment.
- **`start-all.sh` / `stop-all.sh` walk every `package.dtkt.yaml`** under the working directory (including nested ones in `examples/` if any are added). Keep manifests scoped to actual integration roots, or the helpers will try to `dtkt i start -d` against fixtures.
- **`fivetran` lives in `packages/` but isn't in `GO_INTEGRATIONS`.** It exists alongside the other integrations but isn't fanned out by the root Taskfile or wired into [skaffold.yaml](skaffold.yaml). Don't assume `task <verb>` covers it - operate on it directly with its own Taskfile.
- **`packages/openai/` uses OpenAPI codegen** via `oapi-codegen` driven by `generate.go` and `oapigen.yaml` (output to `pkg/oapigen/`). It also keeps a `buf.yaml` for proto linting only - no `buf.gen.yaml`, so `buf generate` would be a no-op there.
- **`packages/bigquery/` requires CGO.** It pins `CGO_ENABLED=1` and `CXX=clang++` in its Taskfile env. Building outside of `task build` (e.g. cross-compiling) needs the same toolchain available.
- **VSCode workspace does not include every package.** [dtkt-integrations.code-workspace](dtkt-integrations.code-workspace) only opens `bigquery`, `metabusiness`, `openai`, and the repo root. Other integrations are still on disk; open them as additional folders or work from the repo root.

Contributor-facing rules - commit conventions (release-please routes by file path, not commit scope), sign-off, branch naming, PR rules - are in [CONTRIBUTING.md](CONTRIBUTING.md).

## Style

[Go style](../dtkt-dev/docs/go/style.md) ·
[project-specific](docs/style.md).

## Testing

[Go testing](../dtkt-dev/docs/go/testing.md) ·
[project-specific](docs/testing.md) (flows are the integration test surface,
per-package Go test layout).
