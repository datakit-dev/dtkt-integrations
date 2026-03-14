# Email Integration

The `Email` integration implements:

- `BaseService`
- `EmailService`
- `EventService`

The `Email` integration allows sending individual, batched, or streamed emails to any compatible SMTP server and subscribing to email events through POP3 and IMAP.

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
dtkt intgr start email --runtime docker
```

This launches the integration in a Docker container.

### Ensure your integration is registered

In another terminal, run:

```shell
dtkt intgr get email
```

### Configure the integration

There are sample configurations in [`examples/configs`](./examples/configs).
If you'd like to create your own, refer to the integration Config proto message [`config.proto`](./proto/dtkt/emailintgr/v1beta1/config.proto) or simply create one interactively when creating a connection.

## Examples

Refer to the [examples](./examples/README.md) for more detailed examples of how to use the `Email` integration package.

## Limitations

While the `EmailService` interface supports managing templates (create, update, delete),
the `Email` integration only implements this part of the interface for testing purposes.
This integration only saves templates ephemerally in memory. For persistent templates it is recommended to define them in the service config which is loaded at runtime or to use an email service provider that supports templates, such as our [Postmark](../postmark) integration.

The only template language supported in this integration is Go template syntax. This means you can use
`{{ .Variable }}` to access variables in the template.

## Related Packages

- [Mailpit](../mailpit): A simple and easy-to-use email testing tool that allows developers to send, receive, and view emails in a development environment.
- [Postmark](../postmark): Send individual, batched, streamed, or templated emails through Postmark's API.

## Legal

Refer to the repository [README's legal documentation](../../README.md).
