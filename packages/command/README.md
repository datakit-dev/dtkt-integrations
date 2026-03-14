# Command Integration

The `Command` integration implements:

- `BaseService`
- `CommandService`

The `Command` integration allows executing commands locally relative to the running integration or remotely over an SSH connection. It supports single command execution, streamed command execution, executing a stream of commands, batch commands, shells, and terminal sessions!

<img alt="execute-command demo with VHS" src="./examples/command-service/execute-command/vhs.gif" width="800" />

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
dtkt intgr start command --runtime docker
```

This launches the integration in a Docker container.

### Ensure your integration is registered

In another terminal, run:

```shell
dtkt intgr get command
```

### Configure the integration

There are sample configurations in [`examples/configs`](./examples/configs).
If you'd like to create your own, refer to the integration Config proto message [`config.proto`](./proto/dtkt/commandintgr/v1beta1/config.proto) or simply create one interactively when creating a connection.

## Examples

Refer to the [examples](./examples/README.md) for more detailed examples of how to use the `Command` integration package.

## Limitations

- Remote non-shell commands over SSH do not support setting environment variables or working directory.

## Related Packages

N/A

## Legal

Refer to the repository [README's legal documentation](../../README.md).
