# dtkt-integrations style

Conventions for writing or changing integration packages. Operational facts (how to run them, skaffold profile lists, root Taskfile fan-out) are in [../AGENTS.md](../AGENTS.md).

For language-level rules:
- Go: [../../dtkt-dev/docs/go/style.md](../../dtkt-dev/docs/go/style.md)

For contributor-facing rules (commit conventions, sign-off, release-please routing, PR/branch naming): [../CONTRIBUTING.md](../CONTRIBUTING.md).

## Use this for that

### Per-package shape

- **Mirror an existing package.** [packages/email/](../packages/email/) is the minimal reference; [packages/bigquery/](../packages/bigquery/) is fuller. Every package has the same surface: `main.go`, `package.dtkt.yaml` (embedded), `pkg/`, `Dockerfile`, `Dockerfile.dockerignore`, `Taskfile.yaml`, `skaffold.yaml`, `infra/kustomize/{base,overlays/{dev,prd}}`.
- **One `go.mod` per package.** Cross-package sharing goes through `dtkt-sdk`, not a `replace` directive.
- **`main.go` is registration only.** Construct via `integrationsdk.NewFS(<embed.FS>, pkg.NewInstance)`, register services via `integrationsdk.RegisterService(...)` / `RegisterManagedActionService(...)` / `RegisterManagedEventService(...)`, then `intgr.Serve()`. Implementation belongs under `pkg/`.

### `Instance` type

- **One per package, in `pkg/instance.go`.** Holds the decoded `Config` proto and any clients/sub-services the integration owns.
- **`CheckAuth` and `Close` are required.** Production integrations implement `CheckAuth` (a permanent `codes.Unimplemented` stub doesn't ship); `Close` releases every client/pool the instance opened.
- **Type parameter is consistent** across `main.go` and `pkg/`. `*pkgv1beta1.Instance` (or whichever versioned variant) flows through generic helpers like `lib.AuditLogEvents[T]()` - keep it stable across files.

### Versioned APIs

- **Keep `pkg/v1beta1/`, `pkg/v1beta2/` side-by-side** when implementing multiple proto API versions. Both versions stay registered until consumers migrate off the older one.

### Taskfile

- **Include `${TASK_PATH}/Taskfile.go.yaml` with `GO_BUILD_BINARY: pkg`** in every package's `Taskfile.yaml`. That's what gives every package the same `setup`/`generate`/`lint`/`build`/`test` surface.
- **Use `task go:lint` for the `lint` target.** Mirror [packages/openai/Taskfile.yaml](../packages/openai/Taskfile.yaml). [packages/bigquery/Taskfile.yaml](../packages/bigquery/Taskfile.yaml) is currently `echo TODO: fix go lint`; that's a known wart, not a pattern.

### Flow YAML

- **Action `call.method` is fully-qualified** (`ai.v1beta1.AgentService.Run`), not bare. The platform routes by service name.

> The "add a flow under [../flows/](../flows/) when adding a new service or RPC"
> rule lives in [testing.md](testing.md) - flows are the integration test
> surface.

### CGO

- **CGO is opt-in and rare.** `packages/bigquery/` is the only CGO integration today (`CGO_ENABLED=1`, `CXX=clang++`); each new CGO package complicates release builds and Docker layers, so weigh that cost before adding one.

## Patterns to add

> Add as the integration surface grows - error-mapping conventions to gRPC `codes.*`, retry/backoff defaults for managed event sources, `package.dtkt.yaml` icon dimensions and naming, when a helper belongs in `pkg/lib/` vs in `dtkt-sdk` shared libs.
