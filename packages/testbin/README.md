# Testbin Integration

The `Testbin` integration implements:

- `BaseService`
- `EchoService`
- `InspectService`

The `Testbin` integration is a gRPC test fixture - the gRPC analog of [httpbin](https://httpbin.org). It exercises every gRPC streaming pattern (unary, server-stream, client-stream, bidi) with controllable payload echoes, error injection via `google.rpc.Status`, latency, and request introspection. Use it to verify that flows handle gRPC mechanics end-to-end - both happy paths and failure paths - without needing a real upstream.

<img alt="echo-unary demo with VHS" src="./examples/echo-service/echo-unary/vhs.gif" width="800" />

## Getting Started

Download or build the [DataKit CLI](https://withdatakit.com/docs/cli/dtkt/overview) and ensure it is accessible to your `$PATH`.

```shell
dtkt version
```

### Start the Package

You can start the integration in one of two ways:

#### Option 1: Develop from Source

Clone down this repository and the [DataKit SDK](https://github.com/datakit-dev/dtkt-sdk) side by side.

From the root of the integration package, run:

```shell
dtkt intgr dev
```

This launches the integration in a development loop that automatically rebuilds and restarts on detected changes.

<img alt="dev demo with VHS" src="./examples/dev/vhs.gif" width="800" />

#### Option 2: Run a prebuilt binary in Docker

Ensure the Docker daemon is running and run the following command:

```shell
dtkt intgr start testbin --runtime docker
```

This launches the integration in a Docker container.

### Ensure your integration is registered

In another terminal, run:

```shell
dtkt intgr get testbin
```

### Configure the integration

There is a sample configuration in [`examples/configs`](./examples/configs).

Testbin has no real configuration - the `Config.payload` is a free-form `google.protobuf.Value` that round-trips through `InspectService.GetConfig`. Use it to verify that connection-level config plumbing works end-to-end.

## Examples

Refer to the [examples](./examples/README.md) for more detailed examples of how to use the `Testbin` integration package.

## Limitations

- Testbin is intended for testing only - do not deploy as a production dependency.
- Multi-valued metadata is preserved, but binary metadata (keys ending in `-bin`) is base64-encoded in the response since the proto representation uses string values.

## Related Packages

N/A

## Legal

Refer to the repository [README's legal documentation](../../README.md).
