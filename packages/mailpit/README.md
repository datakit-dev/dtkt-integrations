# Mailpit Integration

The `Mailpit` integration implements:

- `ActionService`
- `BaseService`
- `EmailService`

Mailpit is a simple and easy-to-use email testing tool that allows developers to send, receive, and view emails in a development environment.
The integration provides actions and event sources allowing you to send emails, subscribe to email events, as well as every
other action that Mailpit's API supports.

This package is meant only for development and testing purposes. It is not intended for production use.
This Mailpit setup does not use TLS/SSL, encryption, or any other security measures.

<img alt="send-email demo with VHS" src="./examples/email-service/send-email/vhs.gif" width="800" />

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
dtkt intgr start mailpit --runtime docker
```

This launches the integration in a Docker container.

### Ensure your integration is registered

In another terminal, run:

```shell
dtkt intgr get mailpit
```

### Configure the integration

There are sample configurations in [`examples/configs`](./examples/configs).
If you'd like to create your own, refer to the Config struct in [pkg/service.go](./pkg/service.go) or simply create one interactively when creating a connection.

## Examples

Refer to the [examples](./examples/README.md) for more detailed examples of how to use the `Mailpit` integration package.

## Limitations

While the `EmailService` interface supports managing templates (create, update, delete),
this `Mailpit` integration only implements this part of the interface for testing purposes.
This integration only saves templates ephemerally in memory. For persistent templates it is recommended to define them in the service config which is loaded at runtime or to use an email service provider that supports templates, such as our [Postmark](../postmark) integration.

The only template language supported in this integration is Go template syntax. This means you can use
`{{ .Variable }}` to access variables in the template.

## Related Packages

- [Email](../email): A simple and easy-to-use email testing tool that allows developers to send, receive, and view emails in a development environment.
- [Postmark](../postmark): Send individual, batched, streamed, or templated emails through Postmark's API.
